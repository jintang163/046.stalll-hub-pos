import os
from dotenv import load_dotenv

load_dotenv()


class Settings:
    CLICKHOUSE_HOST = os.getenv("CLICKHOUSE_HOST", "127.0.0.1")
    CLICKHOUSE_PORT = int(os.getenv("CLICKHOUSE_PORT", "8123"))
    CLICKHOUSE_USER = os.getenv("CLICKHOUSE_USER", "default")
    CLICKHOUSE_PASSWORD = os.getenv("CLICKHOUSE_PASSWORD", "")
    CLICKHOUSE_DATABASE = os.getenv("CLICKHOUSE_DATABASE", "stall_hub_pos")

    SERVER_HOST = os.getenv("SERVER_HOST", "0.0.0.0")
    SERVER_PORT = int(os.getenv("SERVER_PORT", "8010"))

    FORECAST_DAYS = int(os.getenv("FORECAST_DAYS", "3"))
    HISTORY_DAYS = int(os.getenv("HISTORY_DAYS", "90"))
    PROPHET_CHANGES = float(os.getenv("PROPHET_CHANGES", "0.05"))


settings = Settings()
