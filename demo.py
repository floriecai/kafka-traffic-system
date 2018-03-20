import os
import subprocess
import spur

GO_ENV = {
    'GOPATH': '/home/416/go',
    'PATH': '$PATH:/usr/local/go/bin:$GOPATH/bin'
}

FILES = {
    'main': ['node/main.go', '20.36.31.108:12345'],
    'server': ['server/server.go', '-c', 'server/config.json']
}


def get_ip(vm):
    with open('vm-ips.txt') as f:
        for l, v in enumerate(f):
            if l == vm - 1:
                return v.strip()


def get_pw(vm):
    with open('vm-pws.txt') as f:
        for l, v in enumerate(f):
            if l == vm - 1:
                return v.strip()


def handle_action(shell):
    """
    Executes user-input actions on VM

    Valid actions:
        exit, quit: terminate script
        pull: git pull latest commit
        run: go run file
        clean: kill zombie processes
        custom: have it your way
    """
    loop = True
    action = input('What action would you like to do?\n')

    if action == 'exit' or action == 'quit':
        loop = False
    elif action == 'pull':
        # I'd like to have this functionality,
        # but I can't leave my password in a shared repository
        result = shell.run(['git', 'pull'],
                            cwd='proj2_g4w8_g6y9a_i6y8_o5z8',
                            encoding='utf-8')
        print(result.output)
    elif action == 'run':
        fp = input('Which file would you like to run?\n')
        args = ['go', 'run'] + FILES[fp]
        result = shell.run(args,
                            cwd='proj2_g4w8_g6y9a_i6y8_o5z8',
                            update_env=GO_ENV,
                            encoding='utf-8')
        print(result.output)
    elif action == 'clean':
        # only for server so far
        result = shell.run(['fuser', '12345/tcp'],
                            encoding='utf-8')
        print(result.output)
        pid = result.output.strip()
        result = shell.run(['kill', '-9', pid],
                            encoding = 'utf-8')
        print(result.output)
    elif action == 'custom':
        cmd = input('Enter your custom action\n')
        result = shell.run(cmd.split(),
                            encoding='utf-8')
        print(result.output)
    else:
        print('Not a valid action; try again')
    
    return loop


def main():
    """
    Process command line arguments and run appropriate files
    """
    print('CPSC 416 Project 2 General-Purpose Script\n')

    vm = int(input('Which VM would you like to connect to? (1 to 10)\n'))
    
    ip = get_ip(vm)
    pw = get_pw(vm)

    print('Connecting to %s with password %s\n' % (ip, pw))

    shell = spur.SshShell(username='416', hostname=ip, password=pw, connect_timeout=5)
    shell.run(['echo', 'Verifying login'], encoding='utf-8')
    
    print('Successfully connected\n')

    loop = True

    while loop:
        loop = handle_action(shell)


if __name__ == '__main__':
    main()
