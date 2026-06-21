import logging
from datetime import datetime, timedelta
from typing import List, Dict, Optional, Tuple

import pandas as pd

from .config import settings
from .clickhouse_client import get_clickhouse_client

logger = logging.getLogger(__name__)

try:
    from prophet import Prophet
    PROPHET_AVAILABLE = True
except ImportError:
    PROPHET_AVAILABLE = False
    logger.warning("Prophet not available, will use fallback method")


def fetch_sku_daily_sales(
    store_id: int,
    history_days: int,
    sku_ids: Optional[List[int]] = None,
) -> pd.DataFrame:
    client = get_clickhouse_client()
    if client is None:
        logger.error("ClickHouse not available")
        return pd.DataFrame(columns=["ds", "y", "sku_id", "sku_name", "product_id", "product_name"])

    end_date = datetime.now().date()
    start_date = end_date - timedelta(days=history_days)

    query = """
    SELECT
        toString(created_date) as ds,
        SUM(quantity) as y,
        sku_id,
        sku_name,
        product_id,
        product_name
    FROM stall_hub_pos.ch_order_items FINAL
    WHERE store_id = %(store_id)s
      AND created_date >= %(start_date)s
      AND created_date < %(end_date)s
      AND status != -1
    """
    params = {
        "store_id": store_id,
        "start_date": start_date.strftime("%Y-%m-%d"),
        "end_date": end_date.strftime("%Y-%m-%d"),
    }

    if sku_ids:
        placeholders = ", ".join([f"%(sku_{i})s" for i in range(len(sku_ids))])
        query += f" AND sku_id IN ({placeholders})"
        for i, sku_id in enumerate(sku_ids):
            params[f"sku_{i}"] = sku_id

    query += " GROUP BY created_date, sku_id, sku_name, product_id, product_name ORDER BY ds, sku_id"

    try:
        result = client.query(query, parameters=params)
        df = pd.DataFrame(
            result.result_rows,
            columns=["ds", "y", "sku_id", "sku_name", "product_id", "product_name"],
        )
        df["ds"] = pd.to_datetime(df["ds"])
        df["y"] = df["y"].astype(float)
        return df
    except Exception as e:
        logger.error(f"Failed to fetch SKU daily sales: {e}")
        return pd.DataFrame(columns=["ds", "y", "sku_id", "sku_name", "product_id", "product_name"])


def predict_sku_prophet(
    df_sku: pd.DataFrame,
    forecast_days: int,
) -> Optional[Dict]:
    if len(df_sku) < 7:
        return None

    df_train = df_sku[["ds", "y"]].copy()
    df_train = df_train.sort_values("ds").reset_index(drop=True)

    try:
        model = Prophet(
            changepoint_prior_scale=settings.PROPHET_CHANGES,
            yearly_seasonality=False,
            weekly_seasonality=True,
            daily_seasonality=False,
        )
        model.fit(df_train)

        future = model.make_future_dataframe(periods=forecast_days, freq="D")
        forecast = model.predict(future)

        future_forecast = forecast.tail(forecast_days)

        daily_forecast = []
        for _, row in future_forecast.iterrows():
            daily_forecast.append({
                "date": row["ds"].strftime("%Y-%m-%d"),
                "forecast": max(0, round(float(row["yhat"]), 2)),
                "lower": max(0, round(float(row["yhat_lower"]), 2)),
                "upper": max(0, round(float(row["yhat_upper"]), 2)),
            })

        total_forecast = sum(d["forecast"] for d in daily_forecast)
        avg_daily = total_forecast / forecast_days if forecast_days > 0 else 0

        last_7_days = df_train.tail(7)["y"].mean() if len(df_train) >= 7 else df_train["y"].mean()
        avg_before = df_train["y"].mean()
        if avg_before > 0:
            change_pct = (avg_daily - avg_before) / avg_before * 100
        else:
            change_pct = 0

        if change_pct > 5:
            trend = "up"
        elif change_pct < -5:
            trend = "down"
        else:
            trend = "stable"

        return {
            "daily_forecast": daily_forecast,
            "total_forecast": round(total_forecast, 2),
            "avg_daily": round(avg_daily, 2),
            "confidence_lower": round(sum(d["lower"] for d in daily_forecast), 2),
            "confidence_upper": round(sum(d["upper"] for d in daily_forecast), 2),
            "trend": trend,
            "trend_change_pct": round(change_pct, 2),
            "history_days": len(df_train),
            "avg_daily_history": round(float(avg_before), 2),
            "method": "prophet",
        }
    except Exception as e:
        logger.warning(f"Prophet prediction failed, using fallback: {e}")
        return None


def predict_sku_fallback(
    df_sku: pd.DataFrame,
    forecast_days: int,
) -> Dict:
    df_sorted = df_sku.sort_values("ds").reset_index(drop=True)

    if len(df_sorted) >= 28:
        recent = df_sorted.tail(28)
        avg_val = recent["y"].mean()
        std_val = recent["y"].std()
    elif len(df_sorted) >= 7:
        recent = df_sorted.tail(7)
        avg_val = recent["y"].mean()
        std_val = recent["y"].std()
    else:
        avg_val = df_sorted["y"].mean()
        std_val = df_sorted["y"].std() if len(df_sorted) > 1 else avg_val * 0.2

    if pd.isna(std_val):
        std_val = avg_val * 0.2

    weekly_factor = 1.0
    if len(df_sorted) >= 7:
        last_7 = df_sorted.tail(7)["y"].mean()
        prev_7 = df_sorted.tail(14).head(7)["y"].mean() if len(df_sorted) >= 14 else last_7
        if prev_7 > 0:
            weekly_factor = last_7 / prev_7
        weekly_factor = max(0.5, min(1.5, weekly_factor))

    daily_forecast = []
    base_date = datetime.now().date() + timedelta(days=1)
    for i in range(forecast_days):
        day_date = base_date + timedelta(days=i)
        day_of_week = day_date.weekday()

        if day_of_week >= 5:
            weekend_factor = 1.15
        else:
            weekend_factor = 1.0

        forecast_val = max(0, avg_val * weekly_factor * weekend_factor)
        lower_val = max(0, forecast_val - std_val)
        upper_val = forecast_val + std_val

        daily_forecast.append({
            "date": day_date.strftime("%Y-%m-%d"),
            "forecast": round(forecast_val, 2),
            "lower": round(lower_val, 2),
            "upper": round(upper_val, 2),
        })

    total_forecast = sum(d["forecast"] for d in daily_forecast)
    avg_daily = total_forecast / forecast_days if forecast_days > 0 else 0
    avg_history = avg_val

    if avg_history > 0:
        change_pct = (avg_daily - avg_history) / avg_history * 100
    else:
        change_pct = 0

    if change_pct > 5:
        trend = "up"
    elif change_pct < -5:
        trend = "down"
    else:
        trend = "stable"

    return {
        "daily_forecast": daily_forecast,
        "total_forecast": round(total_forecast, 2),
        "avg_daily": round(avg_daily, 2),
        "confidence_lower": round(sum(d["lower"] for d in daily_forecast), 2),
        "confidence_upper": round(sum(d["upper"] for d in daily_forecast), 2),
        "trend": trend,
        "trend_change_pct": round(change_pct, 2),
        "history_days": len(df_sorted),
        "avg_daily_history": round(float(avg_history), 2),
        "method": "fallback_moving_avg",
    }


def predict_sku(
    df_sku: pd.DataFrame,
    forecast_days: int,
) -> Dict:
    if PROPHET_AVAILABLE and len(df_sku) >= 14:
        result = predict_sku_prophet(df_sku, forecast_days)
        if result is not None:
            return result
    return predict_sku_fallback(df_sku, forecast_days)


def generate_store_forecast(
    store_id: int,
    forecast_days: Optional[int] = None,
    history_days: Optional[int] = None,
    sku_ids: Optional[List[int]] = None,
) -> Tuple[List[Dict], float]:
    if forecast_days is None:
        forecast_days = settings.FORECAST_DAYS
    if history_days is None:
        history_days = settings.HISTORY_DAYS

    df_sales = fetch_sku_daily_sales(store_id, history_days, sku_ids)
    if df_sales.empty:
        logger.warning(f"No sales data for store {store_id}")
        return [], 0.0

    sku_groups = df_sales.groupby(["sku_id", "sku_name", "product_id", "product_name"])

    sku_forecasts = []
    prophet_count = 0
    total_skus = 0

    for (sku_id, sku_name, product_id, product_name), df_sku in sku_groups:
        if len(df_sku) < 3:
            continue

        result = predict_sku(df_sku, forecast_days)
        sku_forecasts.append({
            "sku_id": int(sku_id),
            "sku_name": str(sku_name),
            "product_id": int(product_id),
            "product_name": str(product_name),
            **result,
        })
        total_skus += 1
        if result.get("method") == "prophet":
            prophet_count += 1

    sku_forecasts.sort(key=lambda x: x["total_forecast"], reverse=True)

    quality_score = (prophet_count / total_skus) if total_skus > 0 else 0.0

    return sku_forecasts, quality_score
