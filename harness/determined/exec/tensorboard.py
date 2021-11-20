import json
import logging
import os
import subprocess
import sys
import time
from typing import Any, Callable, Dict, List, Tuple

import boto3
import requests

import determined.tensorboard.fetchers

TENSORBOARD_TRIGGER_READY_MSG = "TensorBoard contains metrics"
CONFIG_PATH = "/run/determined/tensorboard/experiment_config.json"
FETCH_INTERVAL = 1
MAX_WAIT_TIME = 600

logger = logging.getLogger("determined.exec.tensorboard")
formatter = logging.Formatter("%(name)s - %(levelname)s - %(message)s")
stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setFormatter(formatter)
logger.addHandler(stderr_handler)


# XXX remove me.
root_logger = logging.getLogger()
root_logger.setLevel(logging.DEBUG)


def get_config() -> Dict[str, Any]:
    # XXX Probably could just get the storage config.
    with open(CONFIG_PATH) as config_file:
        exp_config = json.load(config_file)
    return exp_config


def set_s3_region() -> None:
    # XXX Can get this from the exp_config
    bucket = os.environ.get("AWS_BUCKET")
    if bucket is None:
        return

    endpoint_url = os.environ.get("DET_S3_ENDPOINT_URL", None)
    client = boto3.client("s3", endpoint_url=endpoint_url)
    bucketLocation = client.get_bucket_location(Bucket=bucket)

    region = bucketLocation["LocationConstraint"]

    if region is not None:
        # We have observed that in US-EAST-1 the region comes back as None
        # and if AWS_REGION is set to None, tensorboard fails to pull events.
        print(f"Setting AWS_REGION environment variable to {region}.")
        os.environ["AWS_REGION"] = str(region)


# XXX Need to impliment something like a:
# determined.tensorboard.fetchers.build -> build_storage_fetcher()


def get_tensorboard_args(tb_version, tfevents_dir, add_args):
    """Build tensorboard startup args.

    Args are added and deprecated at the mercy of tensorboard; all of the below are necessary to
    support versions 1.14, 2.4, and 2.5

    - Tensorboard 2+ no longer exposes all ports. Must pass in "--bind_all" to expose localhost
    - Tensorboard 2.5.0 introduces an experimental feature (default load_fast=true)
    which prevents multiple plugins from loading correctly.
    """
    task_id = os.environ["DET_TASK_ID"]
    port = os.environ["TENSORBOARD_PORT"]

    tensorboard_args = [
        "tensorboard",
        f"--port={port}",
        f"--path_prefix=/proxy/{task_id}",
        *add_args,
    ]

    # Version dependant args
    version_parts = tb_version.split(".")
    major = int(version_parts[0])
    minor = int(version_parts[1])

    if major >= 2:
        tensorboard_args.append("--bind_all")
    if major > 2 or major == 2 and minor >= 5:
        tensorboard_args.append("--load_fast=false")

    tensorboard_args.append(f"--logdir={tfevents_dir}")

    return tensorboard_args


def get_tensorboard_url():
    task_id = os.environ["DET_TASK_ID"]
    port = os.environ["TENSORBOARD_PORT"]
    tensorboard_addr = f"http://localhost:{port}/proxy/{task_id}"
    return f"{tensorboard_addr}/data/plugin/scalars/tags"


def check_for_metrics():
    tensorboard_url = get_tensorboard_url()
    tags = {}
    try:
        # Attempt to retrieve metrics from tensorboard.
        res = requests.get(tensorboard_url)
        res.raise_for_status()
        logger.debug(f"requests.get({tensorboard_url}) -> Response.content: {res.content}")
        tags = res.json()

    except (requests.exceptions.ConnectionError, requests.exceptions.HTTPError) as exp:
        logger.warning(f"OK: Tensorboard not responding to HTTP: {exp}")

    except ValueError as exp:
        # XXX TTUCKER, when does a ValueError occur?  raise_for_status()?
        logger.warning(str(exp))

    except json.JSONDecodeError as exp:
        logger.warning(f"OK: Could not JSONDecode Tensorboard HTTP response: {exp}")

    if len(tags) != 0 and any([len(v) for v in tags.values()]):
        print(TENSORBOARD_TRIGGER_READY_MSG)
        return True

    return False


def start_tensorboard(tb_version, paths, add_tb_args):

    config = get_config()

    # Create local temporary directory
    script_dir = os.path.dirname(__file__)
    local_dir = os.path.join(script_dir, "tf_events")
    os.makedirs(local_dir, mode=0o777, exist_ok=True)

    # Get fetcher and perform initial fetch
    fetcher = determined.tensorboard.fetchers.build(config, paths, local_dir)
    fetcher.fetch_new()

    # Build Tensorboard args and launch process.
    tb_args = get_tensorboard_args(tb_version, local_dir, add_tb_args)
    logger.debug(f"tensorboard args: {tb_args}")
    tensorboard_process = subprocess.Popen(tb_args)

    # Loop state.
    stop_time = time.time() + MAX_WAIT_TIME
    tb_has_metrics = False

    try:
        while True:
            # Check if tensorboard process is still alive.
            ret_code = tensorboard_process.poll()
            if ret_code is not None:
                raise RuntimeError(f"Tensorboard process died, exit code({ret_code}).")
            logger.debug("Tensorboard process still alive.")  # XXX Spams logs

            # Check if we have reached a timeout without receiving metrics
            if not tb_has_metrics and time.time() > stop_time:
                raise RuntimeError("We reached the timeout without receiving metrics.")

            if not tb_has_metrics:
                tb_has_metrics = check_for_metrics()

            fetcher.fetch_new()
            time.sleep(FETCH_INTERVAL)

    except Exception as exp:
        logger.error(str(exp))

    finally:
        if tensorboard_process.poll() is None:
            logger.debug("Killing tensorboard process")
            tensorboard_process.kill()

    return tensorboard_process.wait()


if __name__ == "__main__":
    tb_version = sys.argv[1]
    paths = sys.argv[2].split(",")
    additional_tb_args = sys.argv[3:]

    ret = start_tensorboard(tb_version, paths, additional_tb_args)
    sys.exit(ret)
