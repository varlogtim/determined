import boto3  # XXX is there a cost to importing here? Why do the storage_managers not do this?
import datetime
import logging
import os
import urllib.parse
import sys

from typing import Any, Dict, Generator, List, Optional, Tuple
from .base import Fetcher


# XXX There has to be a way to programmatically get the module we are in.
logger = logging.getLogger("determined.tensorboard.fetchers.s3")

# XXX TTUCKER, if we remove the lines below, will we use the 'root' logger stream handler?
formatter = logging.Formatter("%(module)s - %(levelname)s - %(message)s")
stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setFormatter(formatter)
logger.addHandler(stderr_handler)


class S3Fetcher(Fetcher):
    def __init__(self, storage_config: Dict[str, Any], paths: List[str], local_dir: Optional[str]):
        # XXX TTUCKER impl what happens when None is pass for local_dir
        self.client = boto3.client(
            "s3",
            endpoint_url=storage_config.get("endpoint_url", None),
            aws_access_key_id=storage_config.get("access_key", None),
            aws_secret_access_key=storage_config.get("secret_key", None)
        )
        self._tfevents_files = {}

        self.bucket = storage_config.get("bucket")
        self.local_dir = local_dir
        self.paths = paths

    def _find_keys(self, prefix: str) -> Generator[Tuple[str, datetime.datetime], None, None]:
        """Generates tuples of (s3_key, datetime.datetime)."""

        logger.debug(f"Listing keys in bucket '{self.bucket}' with prefix '{prefix}'")

        prefix = urllib.parse.urlparse(prefix).path.lstrip("/")

        list_args = {}
        while True:
            list_dict = self.client.list_objects_v2(
                Bucket=self.bucket,
                Prefix=prefix,
                **list_args,
            )
            logger.debug(f"list_objects_v2 response dict: {list_dict}")

            # XXX indicate whether our return list is empty.
            for s3_obj in list_dict.get("Contents", []):
                yield (s3_obj["Key"], s3_obj["LastModified"])

            list_args.pop("ContinuationToken", None)
            if list_dict["IsTruncated"]:
                list_args["ContinuationToken"] = list_dict["NextContinuationToken"]
                continue

            break

    def fetch_new(self) -> None:
        keys_to_download = []

        # Look at all files in our storage location.
        for path in self.paths:
            logger.debug(f"Looking at path: {path}")

            for key, mtime in self._find_keys(path):
                prev_mtime = self._tfevents_files.get(key, None)

                if prev_mtime is not None and prev_mtime >= mtime:
                    logger.debug(f"Key not new: '{key}'")
                    continue

                logger.debug(f"Found new key: '{key}'")
                keys_to_download.append(key)
                self._tfevents_files[key] = mtime

        # Download the new or updated files.
        for key in keys_to_download:
            local_path = os.path.join(self.local_dir, self.bucket, key)

            dir_path = os.path.dirname(local_path)
            os.makedirs(dir_path, exist_ok=True)

            with open(local_path, "wb+") as local_file:
                self.client.download_fileobj(self.bucket, key, local_file)

            logger.debug(f"Downloaded file to local: {local_path}")
