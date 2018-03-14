import argparse
import os
import subprocess

def ssh(ip):
    """
    Our Azure VMs do not have passwords
    Only log in with SSH key

    TODO: Finish drafting this
    """
    subprocess.run(['ssh', ip])


def main():
    """
    TODO: Finish drafting this
    """
    parser = argparse.ArgumentParser(description='Project 2 demo script')
    parser.add_argument('--vm', metavar='N', type=int, help='the index of the VM to SSH into')
    parser.add_argument('--file', '-f' metavar='file', type=int, help='the file to run')
