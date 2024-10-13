#!/usr/bin/env python3

import socket

SERVER_HOST = '127.0.0.1'
SERVER_PORT = 8080

auth_headers = {
    "ClientID": "second_client",
}

def build_headers(headers):
    """
    Build headers into the correct format to be sent over TCP.
    """
    return "\r\n".join(f"{key}: {value}" for key, value in headers.items()) + "\n"

def start_client():
    # Create a TCP socket
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

    try:
        # Connect to the server
        client_socket.connect((SERVER_HOST, SERVER_PORT))
        print(f"Connected to server at {SERVER_HOST}:{SERVER_PORT}")

        # Send authentication headers to the server
        headers = build_headers(auth_headers)
        client_socket.send(headers.encode('utf-8'))
        print("Authentication headers sent")

        # Main loop to send and receive messages
        while True:
            # Receive message from the server
            response = client_socket.recv(1024).decode('utf-8')
            if response:
                print(f"Server: {response}")

            # Send message to the server
            message = input("You: ")
            client_socket.send(message.encode('utf-8'))

            # If you want to close the connection with a certain message
            if message.lower() == "exit":
                print("Closing connection.")
                break

    except Exception as e:
        print(f"Error: {e}")
    finally:
        # Close the connection
        client_socket.close()

if __name__ == "__main__":
    start_client()

