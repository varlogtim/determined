from typing import Any, Dict, List, Optional

from .azure import AzureFetcher
from .base import Fetcher
from .gcs import GCSFetcher
from .s3 import S3Fetcher

__all__ = [
    "S3Fetcher",
    "GCSFetcher",
    "AzureFetcher",
]

_FETCHERS = {
    "s3": S3Fetcher,
    "gcs": GCSFetcher,
    "azure": AzureFetcher,
}


def build(config: Dict[str, Any], paths: List[str], local_dir: str) -> Fetcher:
    storage_config = config.get("checkpoint_storage", None)
    if storage_config is None:
        raise ValueError("config does not contain a 'checkpoint_stroage' key")

    # XXX Fix me. Read the other code related to this.
    storage_type = storage_config.get("type", None)
    if storage_type is None or storage_type not in _FETCHERS:
        raise ValueError(f"checkpoint_storage type '{storage_type}' is not supported")

    return _FETCHERS[storage_type](storage_config, paths, local_dir)
