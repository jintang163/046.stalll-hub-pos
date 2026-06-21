import logging
from typing import Optional

import clickhouse_connect
from clickhouse_connect.driver.client import Client

from .config import settings

logger = logging.getLogger(__name__)

_clickhouse_client: Optional[Client] = None


def get_clickhouse_client() -> Optional[Client]:
    global _clickhouse_client
    if _clickhouse_client is not None:
        return _clickhouse_client

    try:
        _clickhouse_client = clickhouse_connect.get_client(
            host=settings.CLICKHOUSE_HOST,
            port=settings.CLICKHOUSE_PORT,
            username=settings.CLICKHOUSE_USER,
            password=settings.CLICKHOUSE_PASSWORD,
            database=settings.CLICKHOUSE_DATABASE,
        )
        logger.info("ClickHouse connected successfully")
        return _clickhouse_client
    except Exception as e:
        logger.error(f"Failed to connect to ClickHouse: {e}")
        return None


def test_clickhouse_connection() -> bool:
    try:
        client = get_clickhouse_client()
        if client is None:
            return False
        result = client.query("SELECT 1")
        return result.result_rows[0][0] == 1
    except Exception as e:
        logger.error(f"ClickHouse connection test failed: {e}")
        return False
