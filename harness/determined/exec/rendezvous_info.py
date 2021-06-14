import distutils.util
import os

from determined.common import api
from determined.common.api import certs

if __name__ == "__main__":
    master_addr = os.environ["DET_MASTER_ADDR"]
    master_port = int(os.environ["DET_MASTER_PORT"])
    use_tls = distutils.util.strtobool(os.environ.get("DET_USE_TLS", "false"))
    trial_id = os.environ["DET_TRIAL_ID"]
    container_id = os.environ["DET_CONTAINER_ID"]

    master_url = f"http{'s' if use_tls else ''}://{master_addr}:{master_port}"

    cert = certs.default_load(master_url)

    r = api.post(
        master_url,
        path=f"/api/v1/trials/{trial_id}/rendezvous_info",
        body=container_id,
        cert=cert,
    )

    # just write the rendezvous info to stdout
    print(r.content.decode("utf8"))
