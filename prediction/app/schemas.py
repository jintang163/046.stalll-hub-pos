from pydantic import BaseModel
from datetime import datetime
from typing import List, Optional


class SKUForecastItem(BaseModel):
    sku_id: int
    sku_name: str
    product_id: int
    product_name: str
    daily_forecast: List[dict]
    total_forecast: float
    avg_daily: float
    confidence_lower: float
    confidence_upper: float
    trend: str


class StoreForecastResponse(BaseModel):
    store_id: int
    store_name: Optional[str] = None
    forecast_date: str
    forecast_days: int
    generated_at: str
    sku_forecasts: List[SKUForecastItem]
    total_forecast_sku_count: int
    data_quality_score: float


class ForecastRequest(BaseModel):
    store_id: int
    forecast_days: Optional[int] = None
    history_days: Optional[int] = None
    sku_ids: Optional[List[int]] = None


class HealthResponse(BaseModel):
    status: str
    clickhouse_connected: bool
    prophet_available: bool
    version: str
