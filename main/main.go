package main

import (
	"log"
	"net"
	"time"
)

func main() {
	// use bitcoin p2p protocol
	peerCfg := &peer.Config{
		// User agent name to advertise.
		UserAgentName: "Bitcoin SV",
		// User agent version to advertise.
		UserAgentVersion: "0.2.2",
		ChainParams:      &chaincfg.MainNetParams,
		Services:         0,
		TrickleInterval:  time.Second * 10,
		Listeners: peer.MessageListeners{
			OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
				log.Println("outbound: received version")
				return nil
			},
			OnVerAck: func(p *peer.Peer, msg *wire.MsgVerAck) {
				log.Println("outbound peer has received ack")
			},
			OnReject: func(p *peer.Peer, msg *wire.MsgReject) {
				// panic if get rejected by full node
				log.Println("peer on reject...")
				panic(msg)
			},
			OnInv: func(p *peer.Peer, msg *wire.MsgInv) {
				gdmsg := wire.NewMsgGetData()
				for _, iv := range msg.InvList {
					switch iv.Type {
					case wire.InvTypeBlock:
						err := gdmsg.AddInvVect(iv)
						if err != nil {
							panic(err)
						}
					case wire.InvTypeTx:
						err := gdmsg.AddInvVect(iv)
						if err != nil {
							panic(err)
						}
					}
				}
				if len(gdmsg.InvList) > 0 {
					p.QueueMessage(gdmsg, nil)
				}
			},
			OnTx: func(p *peer.Peer, msg *wire.MsgTx) {
				log.Println(msg.TxHash())
			},
			OnBlock: func(p *peer.Peer, msg *wire.MsgBlock, buf []byte) {
				// 将新块放入channel中缓存
				log.Println("received new block of block hash " + msg.BlockHash().String())
			},
		},
	}
	p, err := peer.NewOutboundPeer(peerCfg, "3.112.178.254:8333")
	if err != nil {
		log.Println("NewOutboundPeer: error " + err.Error())
		panic(err)
	}
	// Establish the connection to the peer address and mark it connected.
	conn, err := net.Dial("tcp", p.Addr())
	if err != nil {
		log.Println("net.Dial: error " + err.Error())
		panic(err)
	}
	p.AssociateConnection(conn)
	defer p.Disconnect()

	// block main thread
	select {}
}
