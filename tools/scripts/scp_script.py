import subprocess
import os, sys
import re

import appdirs
import argparse

def get_parser() -> argparse.ArgumentParser:
    """Handles all the command line arguments when running this as a script
    """
    parser = argparse.ArgumentParser(
        description='''Helper tool to grab checkpoint data from a cluster.
        General strategy is to get experiment number, get checkpoint UUID from experiment, 
        get experiment. Use the det cli tool to get those. det e -> det e lc {experiment number}
        -> {this script} {uuid}'''
    )

    parser.add_argument("uuid", type=str, help="The UUID of the checkpoint to download.")
    parser.add_argument("-m", "--master", metavar="address",
        help="Address of master, defaults to DET_MASTER",
        default=os.environ.get('DET_MASTER'))
    parser.add_argument("-s", "--shell-id", type=str,
        help="The shell to use. If none is provided, one will be generated.",
        default=None)
    parser.add_argument("-o", "--output-dir", type=str,
        help="Desired output directory for the checkpoint.",
        default=os.getcwd())

    return parser

def build_shell() -> str:
    """In the case that the user doesn't supply a shell, we create one and pass
    it to the other functions

    Returns:
        str: The shell ID for new det shell
    """
    print('Creating shell')
    command = ['det','shell','start', '--show-ssh-command', 'exit']
    # Hacky part 1: we run with --show--ssh-command to make the keys persist in cache
    # then exit the shell we create...
    # Point 1, if we use -d (detach shell), we return before we make the keys persist
    # in cache.
    # Point 2, since we're using subprocess, we'll get sucked into the ssh command
    # here if we don't exit after creating shell
    shell = subprocess.Popen(command, shell=False, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    shell_spill = shell.stdout.readlines()[0].decode('utf-8').strip()
    result = re.search('(\(id:[^\(]*?\))', shell_spill)
    # Hacky part 2: if we ran with -d, we just get the ID, and we're done
    # instead, we run a regex to capture (id=shell id) and throw away the parts
    # we don't need, leaving only shell id
    if result:
        result = result.group(0)[5:-1]
    print(f'Created shell with id: {result}')
    return result

def delete_shell(shell_id: str) -> None:
    """Kills the shell that was auto-created. Shouldn't happen if the user supplies
    the shell id themselves.

    Args:
        shell_id (str): The ID of the shell to kill
    """
    print(f'Killing shell with id:  {shell_id}')
    command = ['det','shell','kill',shell_id]
    shell = subprocess.call(command, shell=False, stderr=subprocess.DEVNULL)

def build_ssh_command(shell_id:str, args:argparse.Namespace) -> list:
    """Builds the command that connects to det shell. There are 2 primary components
    to build:
        - The proxy command that allows us to get directly into the shell, ignoring
          the master and agent
        - The identity file that was generated for this shell

    Args:
        shell_id (str): The Unique ID of the det shell, find via `det shell list`
        args (argparse.Namespace): Arguments passed via CLI
            You can get around this by providing a NamedTuple with shell_id and master

    Returns:
        list: The command in list format (easier to use with subprocess)
    """
    proxy_cmd = f'{sys.executable} -m determined.cli.tunnel {args.master} %h'
    unixy_keypath = os.path.join(appdirs.user_cache_dir('determined'), 'shell', shell_id, 'key')
    hostname = shell_id
    username = 'root'

    cmd = [
        'ssh',
        '-o', f'ProxyCommand={proxy_cmd}',
        '-o', 'StrictHostKeyChecking=no',
        '-o', 'IdentitiesOnly=yes',
        '-i', unixy_keypath,
        f'{username}@{hostname}',
    ]
    return cmd

def build_scp_command(shell_id:str, args:argparse.Namespace) -> list:
    """Builds the command that connects to det shell. There are 2 primary components
    to build:
        - The proxy command that allows us to get directly into the shell, ignoring
          the master and agent
        - The identity file that was generated for this shell

    Args:
        shell_id (str): The Unique ID of the det shell, find via `det shell list`
        args (argparse.Namespace): Arguments passed via CLI
            You can get around this by providing a NamedTuple with shell_id and master

    Returns:
        list: The command in list format (easier to use with subprocess)
    """
    proxy_cmd = f'{sys.executable} -m determined.cli.tunnel {args.master} %h'
    unixy_keypath = os.path.join(appdirs.user_cache_dir('determined'), 'shell', shell_id, 'key')
    hostname = shell_id
    username = 'root'

    cmd = [
        'scp', '-r',
        '-o',  f'ProxyCommand={proxy_cmd}',
        '-o',  'StrictHostKeyChecking=no',
        '-o',  'IdentitiesOnly=yes',
        '-i',  unixy_keypath,
        f'{username}@{hostname}:~/checkpoints/{args.uuid}/',
        args.output_dir
    ]
    return cmd

def copy_checkpoint_from_storage_to_proxy(shell_id:str, args:argparse.Namespace) -> None:
    """SSH into the det shell and run the det CLI to get the checkpoint

    Args:
        shell_id (str): _description_
        args (argparse.Namespace): Arguments passed via CLI
    """
    command = build_ssh_command(shell_id, args)
    print(f'Copying checkpoint {args.uuid} to proxy')
    print(f'Using command {command}')
    download_command = f'det c download {args.uuid}'
    command.append(download_command)
    
    ssh = subprocess.Popen(command, shell=False, stdout=subprocess.PIPE)
    ssh.wait(20) # Wait for the transfer to complete, or give up after 20s
    print(f'Proxy now contains checkpoint data for {args.uuid}')

def scp_checkpoint_from_proxy_to_local(shell_id:str, args:argparse.Namespace) -> None:
    """Use SCP to grab a file from the det shell container

    Args:
        shell_id (str): The shell id, either supplied from user or auto-generated
        args (argparse.Namespace): Arguments passed via CLI
    """
    command = build_scp_command(shell_id, args)
    print('Copying data from proxy to local')
    print(f'Using command {command}')
    scp = subprocess.Popen(command, shell=False)
    scp.wait(20) # Wait for the transfer to complete, or give up after 20s
    print(f'Copied data to {args.output_dir}')

if __name__ == '__main__':
    args = get_parser().parse_args()

    shell_id = args.shell_id if args.shell_id else build_shell()
    copy_checkpoint_from_storage_to_proxy(shell_id, args)
    scp_checkpoint_from_proxy_to_local(shell_id, args)
    if args.shell_id is None:
        delete_shell(shell_id)
