require 'rubygems'
require 'net/ssh'
require 'thread'

# GLOBALS
@USERNAME = "416"
@SERVER_IP_PORT = "20.36.31.108:12345"
@VM_INDEX

# DEMO METHODS

# GO
def go_run_server(ip, pw)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Running server"
		ssh.exec!("source ~/.profile && cd proj2_g4w8_g6y9a_i6y8_o5z8; go run server/server.go -c server/config.json") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
end

def go_run_node(ip, pw)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Running node"
		ssh.exec!("source ~/.profile && cd ~/proj2_g4w8_g6y9a_i6y8_o5z8; go run node/node.go #{@SERVER_IP_PORT}") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
end

# GIT
def git_status(ip, pw)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Printing out status"
		ssh.exec!("source ~/.profile && cd proj2_g4w8_g6y9a_i6y8_o5z8; git status") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
end

def git_checkout(ip, pw, branch)
	Net::SSH.start(ip, @USERNAME, :password => pw) do |ssh|
		puts "Checking out branch #{branch}"
		ssh.exec!("source ~/.profile && cd proj2_g4w8_g6y9a_i6y8_o5z8; git checkout #{branch}") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
end

def git_pull(ip, pw)
	Net::SSH.start(host, @USERNAME, :password => pw) do |ssh|
		puts "Pulling from Stash"
		ssh.exec!("source ~/.profile && cd proj2_g4w8_g6y9a_i6y8_o5z8; git pull") do
			|ch, stream, line|
			puts line
		end
		ssh.close()
	end
end

# PROMPT METHODS
def vm_prompt()
	puts "Which VM would you like to run on?"
	input = gets.chomp
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
			vm_prompt()
		end
	end
end

def start_prompt(ip, pw)
	puts "Choose an option:"
	puts "  (1) Run Go files"
	puts "  (2) Git"
	input = gets.chomp
	case input
	when "1"
		go_prompt(ip, pw)
	when "2"
		git_prompt(ip, pw)
	when "q"
		exit
	else
		puts "Invalid input. Please try again"
		start_prompt()
	end
end

def go_prompt(ip, pw)
	puts "Which file would you like to run?"
	puts "  (1) Server"
	puts "  (2) Node"
	input = gets.chomp
	case input
	when "q"
		exit
	when "1"
		go_run_server(ip, pw)
	when "2"
		go_run_node(ip, pw)
	else
		puts ("Invalid input. Please try again")
		go_prompt()
	end
end

def git_prompt(ip, pw)
	puts "Which Git command would you like to run?"
	puts "  (1) Checkout"
	puts "  (2) Pull"
	input = gets.chomp
	case input
	when "q"
		exit
	when "1"
		puts "Which branch would you like to checkout?"
		branch = gets.chomp
		git_checkout(ip, pw, branch)
	when "2"
		git_pull(ip, pw)
	else
		puts ("Invalid input. Please try again")
		git_prompt()
	end
end

# MAIN METHOD
puts "CPSC416 Project 2 Demo Script"
puts "Stop with 'q' at any time"

vm_prompt()

ip = IO.readlines("vm-ips.txt")[@VM_INDEX].strip
pw = IO.readlines("vm-pws.txt")[@VM_INDEX].strip

start_prompt(ip, pw)
