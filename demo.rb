require 'rubygems'
require 'net/ssh'
require 'thread'

# INSTANCE VARIABLES
@USERNAME = "416"
@RESOURCE_GROUP = "416p2"
@SERVER_IP_PORT = "20.36.31.108:12345"
@VM_INDEX = "-1"

# DEMO METHODS

# TEST
def echo(ip, pw)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		ssh.exec!("echo 'Hello World!'") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end


# AZURE
def az_vm_list()
	puts "Listing VMs"
	# If anyone has a better way of parsing computerName please change this
	system("az vm list -d -g #{@RESOURCE_GROUP} --query \"[?powerState=='VM running']\""\
	       " | grep 'computerName' | grep -o 'vm[0-9]\\+'")
	puts "Finished listing VMs"
	puts ""
end


def az_vm_start()
	puts "Starting VM #{@VM_INDEX}"
	system("az vm start --resource-group #{@RESOURCE_GROUP} --name vm#{@VM_INDEX}")
	puts "VM started"
	puts ""
end


def az_vm_stop()
	puts "Stopping VM #{@VM_INDEX}"
	system("az vm stop --resource-group #{@RESOURCE_GROUP} --name vm#{@VM_INDEX}")
	puts "VM stopped"
	puts ""
end


def az_vm_dealloc()
	puts "Deallocating VM #{@VM_INDEX}"
	system("az vm deallocate --resource-group #{@RESOURCE_GROUP} --name vm#{@VM_INDEX}")
	puts "VM deallocated"
	puts ""
end


# GO
def go_run_server(ip, pw)
	Net::SSH.start(ip, @USERNAME, password: pw) do |ssh|
		puts "Running server"
		ssh.exec!("source ~/.profile && cd proj2_g4w8_g6y9a_i6y8_o5z8;"\
			      "go run server/server.go -c server/config.json") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end


def go_run_node(ip, pw)
	Net::SSH.start(ip, @USERNAME, password: pw) do |ssh|
		puts "Running node"
		ssh.exec!("source ~/.profile && cd ~/proj2_g4w8_g6y9a_i6y8_o5z8;"\
				  "go run node/node.go 20.36.31.108:12345") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end

# GIT
def git_status(ip, pw)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Printing out status"
		ssh.exec!("cd proj2_g4w8_g6y9a_i6y8_o5z8; git status") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end


def git_checkout(ip, pw, branch)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Checking out branch #{branch}"
		ssh.exec!("cd proj2_g4w8_g6y9a_i6y8_o5z8; git checkout #{branch}") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end


def git_pull(ip, pw)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Pulling from Stash"
		ssh.exec!("cd proj2_g4w8_g6y9a_i6y8_o5z8; git pull") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end


def git_add(ip, pw, f)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Adding to commit"
		ssh.exec!("cd proj2_g4w8_g6y9a_i6y8_o5z8; git add #{f}") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end


def git_commit(ip, pw, msg)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Committing"
		ssh.exec!("cd proj2_g4w8_g6y9a_i6y8_o5z8; git commit -m #{msg}") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end


def git_push(ip, pw)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Pushing to Stash"
		ssh.exec!("cd proj2_g4w8_g6y9a_i6y8_o5z8; git push") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
	puts ""
end


# PROMPT METHODS
def vm_prompt()
	puts "What VM operation would you like to perform?"
	puts "(1) Select a VM to work with"
	puts "(2) Start a VM"
	puts "(3) Stop a VM"
	puts "(4) List all active VMs"
	input = gets.chomp
	puts ""
	case input
	when "q"
		exit
	when "1"
		vm_index_prompt()
	when "2"
		puts "Which VM would you like to start?"
		@VM_INDEX = gets.chomp
		az_vm_start()
	when "3"
		puts "Which VM would you like to stop?"
		@VM_INDEX = gets.chomp
		az_vm_stop()
		az_vm_dealloc()
	when "4"
		az_vm_list()
	else
		puts "Invalid argument. Please try again"
		puts ""
		vm_prompt()
	end
end


def vm_index_prompt()
	puts "Which VM would you like to work with?"
	input = gets.chomp
	puts ""
	case input
	when "q"
		exit
	else
		begin
			vm = Integer(input)
			if vm < 1 || vm > 10
				raise ArgumentError.new("VM must be in 1..10")
			end
			@VM_INDEX = vm
		rescue ArgumentError
			puts "Invalid argument. Please try again"
			puts ""
			vm_index_prompt()
		end
	end
end


def start_prompt(ip, pw)
	puts "Choose an option:"
	puts "(1) Run Go files"
	puts "(2) Git"
	puts "(3) Test"
	input = gets.chomp
	puts ""
	case input
	when "1"
		go_prompt(ip, pw)
	when "2"
		git_prompt(ip, pw)
	when "3"
		echo(ip, pw)
	when "q"
		exit
	else
		puts "Invalid input. Please try again"
		puts ""
		start_prompt()
	end
end


def go_prompt(ip, pw)
	puts "Which file would you like to run?"
	puts "(1) Server"
	puts "(2) Node"
	input = gets.chomp
	case input
	when "q"
		exit
	when "1"
		go_run_server(ip, pw)
	when "2"
		go_run_node(ip, pw)
	else
		puts "Invalid input. Please try again"
		puts ""
		go_prompt()
	end
end


def git_prompt(ip, pw)
	puts "Which Git command would you like to run?"
	puts "(1) Status"
	puts "(2) Checkout"
	puts "(3) Pull"
	puts "(4) Add"
	puts "(5) Commit"
	puts "(6) Push"
	input = gets.chomp
	puts ""
	case input
	when "q"
		exit
	when "1"
		git_status(ip, pw)
	when "2"
		puts "Which branch would you like to checkout?"
		branch = gets.chomp
		puts ""
		git_checkout(ip, pw, branch)
	when "3"
		git_pull(ip, pw)
	when "4"
		puts "What file would you like to add? (. is OK)"
		f = gets.chomp
		puts ""
		git_add(ip, pw, f)
	when "5"
		puts "Please enter a commit message"
		msg = gets.chomp
		puts ""
		git_commit(ip, pw, msg)
	when "6"
		git_push(ip, pw)
	else
		puts "Invalid input. Please try again"
		puts ""
		git_prompt()
	end
end


# LOOP METHODS
def az_loop()
	finished = false
	until finished
		vm_prompt()
		puts "Would you like to run more Azure CLI commands? (y/n)"
		input = gets.chomp
		puts ""
		case input
		when "q"
			exit
		when "n"
			finished = true
		end
	end
end


# MAIN
puts "CPSC416 Project 2 Demo Script"
puts "Stop with 'q' at any time"
puts ""

# Start-up
az_loop()

if @VM_INDEX == "-1" || @VM_INDEX.class == "String"
	vm_index_prompt()
else
	ip = IO.readlines("vm-ips.txt")[@VM_INDEX].strip
	pw = IO.readlines("vm-pws.txt")[@VM_INDEX].strip
end

# Main interactive method
start_prompt(ip, pw)

# Clean-up
az_loop()
