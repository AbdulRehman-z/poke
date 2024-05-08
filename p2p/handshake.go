package p2p

type HandshakeFunc func(*TCPPeer, GameVariant, string, string) (*HandshakePass, error)
