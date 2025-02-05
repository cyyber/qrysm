# Simple Local Testnet

These scripts allow for running a small local testnet with a default of 4 beacon nodes, 4 validator clients and 4 gzond execution clients using Kurtosis.
This setup can be useful for testing and development.

## Installation

1. Install [Docker](https://docs.docker.com/get-docker/). Verify that Docker has been successfully installed by running `sudo docker run hello-world`. 

1. Install [Kurtosis](https://docs.kurtosis.com/install/). Verify that Kurtosis has been successfully installed by running `kurtosis version` which should display the version.

1. Install [yq](https://github.com/mikefarah/yq). If you are on Ubuntu, you can install `yq` by running `snap install yq`.

### Apple M1/M2/M3

While doing the docker load step on the latest versions of docker, you might run into this error:

```
lsetxattr com.apple.provenance /manifest.json: operation not supported
```

Follow the steps in the link below:

https://github.com/moby/moby/issues/47517#issuecomment-2536322079

and create an alias for gnu-tar:

```
alias tar="/opt/homebrew/opt/gnu-tar/libexec/gnubin/tar"
```

## Starting the testnet

To start a testnet, from the Qrysm root repository:
```bash

cd ./scripts/local_testnet
./start_local_testnet.sh
```

You will see a list of services running and "Started!" at the end. 
You can also select your own Qrysm docker image to use by specifying it in `network_params.yml` under the `cl_image` key.
Full configuration reference for kurtosis is specified [here](https://github.com/theQRL/zond-package?tab=readme-ov-file#configuration).

To view all running services:

```bash
kurtosis enclave inspect local-testnet
```

To view the logs:

```bash
kurtosis service logs local-testnet $SERVICE_NAME
```

where `$SERVICE_NAME` is obtained by inspecting the running services above. For example, to view the logs of the first beacon node, validator client and gzond:

```bash
kurtosis service logs local-testnet -f cl-1-qrysm-gzond 
kurtosis service logs local-testnet -f vc-1-gzond-qrysm
kurtosis service logs local-testnet -f el-1-gzond-qrysm
```

If you would like to save the logs, use the command:

```bash
kurtosis dump $OUTPUT_DIRECTORY
```

This will create a folder named `$OUTPUT_DIRECTORY` in the present working directory that contains all logs and other information. If you want the logs for a particular service and saved to a file named `logs.txt`:

```bash
kurtosis service logs local-testnet $SERVICE_NAME -a > logs.txt
```
where `$SERVICE_NAME` can be viewed by running `kurtosis enclave inspect local-testnet`.

Some testnet parameters can be varied by modifying the `network_params.yaml` file. Kurtosis also comes with a web UI which can be open with `kurtosis web`.

## Stopping the testnet

To stop the testnet, from the Qrysm root repository:

```bash
cd ./scripts/local_testnet
./stop_local_testnet.sh
```

You will see "Local testnet stopped." at the end. 

## CLI options

The script comes with some CLI options, which can be viewed with `./start_local_testnet.sh --help`. One of the CLI options is to avoid rebuilding Qrysm each time the testnet starts, which can be configured with the command:

```bash
./start_local_testnet.sh -b false
```