import datetime
import logging
import os
import urllib

from typing import Any, Dict, Generator, List, Optional, Tuple

from .base import Fetcher


logger = logging.getLogger("determined.tensorboard.fetchers.azure")


class AzureFetcher(Fetcher):
    def __init__(self, storage_config: Dict[str, Any], paths: List[str], local_dir: Optional[str]):
        # XXX TTUCKER impl what happens when None is pass for local_dir
        # import azure.core.exceptions   # XXX use this?
        from azure.storage import blob

        connection_string = storage_config.get("connection_string", None)
        container = storage_config.get("container", None)
        account_url = storage_config.get("account_url", None)
        credential = storage_config.get("credential", None)

        # XXX This condition from aszure_client is strange. Can we really have neither?
        if storage_config.get("connection_string", None):
            self.client = blob.BlobServiceClient.from_connection_string(connection_string)
        elif account_url:
            self.client = blob.BlobServiceClient(account_url, credential)

        self.container = container if not container.endswith("/") else container[:-1]

        self._tfevents_files = {}

        self.local_dir = local_dir
        self.paths = paths

    def _list(self, prefix: str) -> Generator[Tuple[str, datetime.datetime], None, None]:
        container = self.client.get_container_client(self.container)
        logger.debug(f"Listing keys in container '{self.container}' with '{prefix}'")

        prefix = urllib.parse.urlparse(prefix).path.lstrip("/")

        blobs = container.list_blobs(name_starts_with=prefix)
        logger.debug(f"list_blobs response obj: {blobs}")
        for blob in blobs:
            yield (blob["name"], blob["last_modified"])

    def fetch_new(self) -> None:
        files_to_download = []

        # Look at all files in our storage location.
        for path in self.paths:
            logger.debug(f"Looking at path: {path}")

            for name, mtime in self._list(path):
                prev_mtime = self._tfevents_files.get(name, None)

                if prev_mtime is not None and prev_mtime >= mtime:
                    logger.debug(f"Blob not new: '{name}'")
                    continue

                logger.debug(f"Found new blob: '{name}'")
                files_to_download.append(name)
                self._tfevents_files[name] = mtime

        # Download the new or updated files.
        for name in files_to_download:
            local_path = os.path.join(self.local_dir, self.container, name)

            dir_path = os.path.dirname(local_path)
            os.makedirs(dir_path, exist_ok=True)

            with open(local_path, "wb+") as local_file:
                # XXX wow.... this is the only line that is different.
                stream = self.client.get_blob_client(self.container, name).download_blob()
                stream.readinto(local_file)

            logger.debug(f"Downloaded file to local: {local_path}")

        return
