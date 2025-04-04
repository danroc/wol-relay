# WOL Relay

How it works:

1. The relay listens for WOL packets on all interfaces. It keeps track of all
   addresses of a given network interface.

2. When a WOL packet is received, it checks the sender's address to determine
   which network originated it. Then sends the packet too the broadcast address
   of all other networks.
