#!/usr/bin/env python3

import socket
import threading
import argparse
import json

SERVER_HOST = '127.0.0.1'
SERVER_PORT = 8080

def build_headers(client_id):
    """
    Build headers into the correct format to be sent over TCP.
    """
    headers = {
        "ClientID": client_id,
    }
    return "\r\n".join(f"{key}: {value}" for key, value in headers.items()) + "\n"

def receive_messages(client_socket):
    """
    Continuously receive messages from the server.
    """
    while True:
        try:
            response = client_socket.recv(1024).decode('utf-8')
            if response:
                print(f"Server: {response}")
            else:
                print("Server closed the connection.")
                break
        except Exception as e:
            print(f"Error receiving data: {e}")
            break

def send_messages(client_socket):
    """
    Continuously send messages to the server.
    """
    while True:
        message = input("")
        client_socket.send(message.encode('utf-8'))

        # Optionally exit on 'exit' command
        if message.lower() == "exit":
            print("Closing connection.")
            client_socket.close()
            break

def start_client(client_id):
    # Create a TCP socket
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

    try:
        # Connect to the server
        client_socket.connect((SERVER_HOST, SERVER_PORT))
        print(f"Connected to server at {SERVER_HOST}:{SERVER_PORT}")

        # Send authentication headers to the server
        headers = build_headers(client_id)
        client_socket.send(headers.encode('utf-8'))
        print("Authentication headers sent")

        response = client_socket.recv(1024).decode('utf-8')
        print(f"Server Response: {response}")

        subscribe = {
            "type": "SUB",
            "topic": "temp/#"
        }

        client_socket.send(json.dumps(subscribe).encode('utf-8'))
        print("Subscribe message sent")

        response = client_socket.recv(1024).decode('utf-8')
        print(f"Server Response: {response}")

        # Start threads for sending and receiving
        receive_thread = threading.Thread(target=receive_messages, args=(client_socket,))
        send_thread = threading.Thread(target=send_messages, args=(client_socket,))

        # Start both threads
        receive_thread.start()
        send_thread.start()

        # Wait for both threads to finish
        receive_thread.join()
        send_thread.join()

    except Exception as e:
        print(f"Error: {e}")
    finally:
        # Close the connection
        client_socket.close()

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='TCP Client')
    parser.add_argument('client_id', type=str, help='Client ID for authentication')
    args = parser.parse_args()

    start_client(args.client_id)
