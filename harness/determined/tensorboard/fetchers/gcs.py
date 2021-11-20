import datetime
import logging
import os
import sys
import urllib

import google.cloud.storage

from typing import Any, Dict, Generator, Optional, List, Tuple
from .base import Fetcher


logger = logging.getLogger("determined.tensorboard.fetchers.gcs")
formatter = logging.Formatter("%(name)s - %(levelname)s - %(message)s")
stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setFormatter(formatter)
logger.addHandler(stderr_handler)


class GCSFetcher(Fetcher):
    def __init__(self, storage_config: Dict[str, Any], paths: List[str], local_dir: Optional[str]):
        # XXX TTUCKER impl what happens when None is pass for local_dir

        credentials_file = storage_config.get("credentials_file", None)
        if credentials_file is not None:
            credentials_file_path = os.path.join(os.getcwd(), "gcs_credentials.json")
            with open(credentials_file_path, "w") as file:
                file.write(credentials_file)
            os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = credentials_file_path
            logger.info(f"TTUCKER: wrote credentials_file: {credentials_file}")

        self.client = google.cloud.storage.Client()

        self._tfevents_files = {}

        self.bucket = storage_config.get("bucket")
        self.local_dir = local_dir
        self.paths = paths

    def _list(self, prefix: str) -> Generator[Tuple[str, datetime.datetime], None, None]:

        logger.debug(f"Listing keys in bucket '{self.bucket}' with '{prefix}'")

        prefix = urllib.parse.urlparse(prefix).path.lstrip("/")

        blobs = self.client.list_blobs(self.bucket, prefix=prefix)
        logger.debug(f"list_blobs response obj: {blobs}")

        # XXX indicate whether our return list is empty.
        for blob in blobs:
            yield (blob.name, blob.updated)

    def fetch_new(self) -> None:
        files_to_download = []
        bucket = self.client.bucket(self.bucket)

        # Look at all files in our storage location.
        for path in self.paths:
            logger.debug(f"Looking at path: {path}")

            for name, mtime in self._list(path):
                prev_mtime = self._tfevents_files.get(name, None)

                if prev_mtime is not None and prev_mtime >= mtime:
                    logger.debug(f"File not new: '{name}'")
                    continue

                logger.debug(f"Found new key: '{name}'")
                files_to_download.append(name)
                self._tfevents_files[name] = mtime

        # Download the new or updated files.
        for name in files_to_download:
            local_path = os.path.join(self.local_dir, self.bucket, name)

            dir_path = os.path.dirname(local_path)
            os.makedirs(dir_path, exist_ok=True)

            # XXX wow.... this is the only line that is different.
            bucket.blob(name).download_to_filename(local_path)

            logger.debug(f"Downloaded file to local: {local_path}")

        return
