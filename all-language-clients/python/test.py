from socket import *

from protocol import TCPXProtocol, TCPXMessage


def test_json():
    # socket instance
    tcp_client_socket = socket(AF_INET, SOCK_STREAM)

    # socket connect
    tcp_client_socket.connect(('10.0.203.69', 7171))

    # message
    message = TCPXMessage()
    message.id = 5
    message.header = {
        'header': '/tcpx/client1'
    }
    message.body = 'hello'

    # tcpx instance
    tcpx_protocol = TCPXProtocol('json')

    # tcpx pack
    packed_data = tcpx_protocol.pack(message)

    # socket send
    tcp_client_socket.send(packed_data)
    # socket receive
    receive_data = tcp_client_socket.recv(1024)

    # tcpx unpack
    message = tcpx_protocol.unpack(receive_data)
    print(message.id, message.header, message.body)

def test_protobuf():
    pass

test_json()
