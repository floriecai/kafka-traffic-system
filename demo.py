import argparse
import os
import subprocess

VALID_FILES = {'main': 'node/main.go', 'server': 'server/server.go'}

def run(fp):
    """
    Run a specific file once connecting to a VM
    TODO: Finish drafting this
    """
    subprocess.run(['go', 'run', fp])


def ssh(ip, pw):
    """
    Our Azure VMs do not have passwords
    Only log in with SSH key
    TODO: Finish drafting this
    """
    subprocess.run(['sshpass', '-p', pw, 'ssh', '-i', '~/.ssh/id_rsa', ip])


def main():
    """
    Process command line arguments
    TODO: Finish drafting this
    """
    parser = argparse.ArgumentParser(description='Project 2 demo script')
    parser.add_argument('-v', '--vm', metavar='N', type=int,
                        help='the index of the VM to SSH into')
    parser.add_argument('-f', '--file', metavar='file', type=str,
                        help='the name of the file to run')
    args = parser.parse_args()

    if args.vm < 1 or args.vm > 10:
        raise ValueError('VM must be from 1 to 10')
    if args.file not in VALID_FILES:
        raise ValueError('File must be one of %s' % VALID_FILES)
    
    ip = "416@"

    with open('vm-ips.txt') as f:
        for l, v in enumerate(f):
            if l == args.vm - 1:
                ip += v

    pw = ""

    with open('vm-pws.txt') as f:
        for l, v in enumerate(f):
            if l == args.vm - 1:
                pw = v
    
    fp = VALID_FILES[args.file]

    ssh(ip, pw)
    run(fp)


if __name__ == '__main__':
    main()
