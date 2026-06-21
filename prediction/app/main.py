import logging
from datetime import datetime
from typing import Optional

from fastapi import FastAPI, HTTPException, Query

from .config import settings
from .schemas import (
    StoreForecastResponse,
    ForecastRequest,
    HealthResponse,
    SKUForecastItem,
)
from .predictor import generate_store_forecast, PROPHET_AVAILABLE
from .clickhouse_client import test_clickhouse_connection

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Stall Hub POS - Sales Forecast API",
    description="Prophet-based sales forecasting for SKUs",
    version="1.0.0",
)


@app.get("/health", response_model=HealthResponse)
def health_check():
    ch_ok = test_clickhouse_connection()
    return HealthResponse(
        status="ok" if ch_ok else "degraded",
        clickhouse_connected=ch_ok,
        prophet_available=PROPHET_AVAILABLE,
        version="1.0.0",
    )


@app.post("/forecast/store/{store_id}", response_model=StoreForecastResponse)
def forecast_store(
    store_id: int,
    request: Optional[ForecastRequest] = None,
):
    if request is None:
        forecast_days = settings.FORECAST_DAYS
        history_days = settings.HISTORY_DAYS
        sku_ids = None
    else:
        forecast_days = request.forecast_days or settings.FORECAST_DAYS
        history_days = request.history_days or settings.HISTORY_DAYS
        sku_ids = request.sku_ids

    try:
        sku_forecasts, quality_score = generate_store_forecast(
            store_id=store_id,
            forecast_days=forecast_days,
            history_days=history_days,
            sku_ids=sku_ids,
        )

        sku_items = [
            SKUForecastItem(
                sku_id=sf["sku_id"],
                sku_name=sf["sku_name"],
                product_id=sf["product_id"],
                product_name=sf["product_name"],
                daily_forecast=sf["daily_forecast"],
                total_forecast=sf["total_forecast"],
                avg_daily=sf["avg_daily"],
                confidence_lower=sf["confidence_lower"],
                confidence_upper=sf["confidence_upper"],
                trend=sf["trend"],
            )
            for sf in sku_forecasts
        ]

        return StoreForecastResponse(
            store_id=store_id,
            forecast_date=datetime.now().strftime("%Y-%m-%d"),
            forecast_days=forecast_days,
            generated_at=datetime.now().isoformat(),
            sku_forecasts=sku_items,
            total_forecast_sku_count=len(sku_items),
            data_quality_score=round(quality_score, 2),
        )
    except Exception as e:
        logger.error(f"Forecast failed for store {store_id}: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Forecast failed: {str(e)}")


@app.get("/forecast/store/{store_id}", response_model=StoreForecastResponse)
def forecast_store_get(
    store_id: int,
    forecast_days: Optional[int] = Query(None, description="Number of days to forecast"),
    history_days: Optional[int] = Query(None, description="Number of history days to use"),
):
    try:
        sku_forecasts, quality_score = generate_store_forecast(
            store_id=store_id,
            forecast_days=forecast_days or settings.FORECAST_DAYS,
            history_days=history_days or settings.HISTORY_DAYS,
        )

        sku_items = [
            SKUForecastItem(
                sku_id=sf["sku_id"],
                sku_name=sf["sku_name"],
                product_id=sf["product_id"],
                product_name=sf["product_name"],
                daily_forecast=sf["daily_forecast"],
                total_forecast=sf["total_forecast"],
                avg_daily=sf["avg_daily"],
                confidence_lower=sf["confidence_lower"],
                confidence_upper=sf["confidence_upper"],
                trend=sf["trend"],
            )
            for sf in sku_forecasts
        ]

        return StoreForecastResponse(
            store_id=store_id,
            forecast_date=datetime.now().strftime("%Y-%m-%d"),
            forecast_days=forecast_days or settings.FORECAST_DAYS,
            generated_at=datetime.now().isoformat(),
            sku_forecasts=sku_items,
            total_forecast_sku_count=len(sku_items),
            data_quality_score=round(quality_score, 2),
        )
    except Exception as e:
        logger.error(f"Forecast failed for store {store_id}: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Forecast failed: {str(e)}")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host=settings.SERVER_HOST,
        port=settings.SERVER_PORT,
        reload=True,
    )
