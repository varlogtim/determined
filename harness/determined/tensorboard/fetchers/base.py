import abc

from typing import Any, Dict, List


class Fetcher(metaclass=abc.ABCMeta):
    """Abstract base class for storage fetchers.

    Responsible for fetching tensorflow events files from storage to a local temp directory.
    """

    @abc.abstractmethod
    def __init__(self, storage_config: Dict[str, Any], paths: List[str], local_dir: str):
        pass

    @abc.abstractmethod
    def fetch_new(self) -> None:
        pass
