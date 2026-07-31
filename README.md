# p2p-chat

This is a UDP peer-to-peer chat application which stores known clients within a peer discovery server's hash table.
Once at least two peers are registered, the server bootstraps the two peer clients to direct packets towards one another
over UDP, thus establishing peer-to-peer connection.

Note: Server address hardcoded, update whenever necessary.