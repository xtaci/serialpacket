# serialpacket

SerialPacket is a net.PacketConn implementation over RS232.

It's designed to work with [kcp-go](https://github.com/xtaci/kcp-go) to provide reliable transmission over [LoRa](https://en.wikipedia.org/wiki/LoRa) or other noisy channels.

Test:
```
$  socat -d -d pty,raw,echo=0 pty,raw,echo=0
2022/03/27 22:48:28 socat[14099] N PTY is /dev/pts/5
2022/03/27 22:48:28 socat[14099] N PTY is /dev/pts/6
2022/03/27 22:48:28 socat[14099] N starting data transfer loop with FDs [5,5] and [7,7]

$ export PORT1="/dev/pts/5"
$ export PORT1="/dev/pts/6"
```
