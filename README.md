bsvd
====
[![Build Status](https://travis-ci.org/metasv/bsvd.png?branch=master)](https://travis-ci.org/metasv/bsvd)
[![Go Report Card](https://goreportcard.com/badge/github.com/metasv/mvcd)](https://goreportcard.com/report/github.com/metasv/mvcd)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/metasv/mvcd)

bsvd is a full node Bitcoin (BSV) implementation written in Go (golang).

This project is a port of the [bchd](https://github.com/gcash/bchd) codebase to Bitcoin (BSV). It provides a high powered
and reliable blockchain server which makes it a suitable backend to serve blockchain data to lite clients and block explorers
or to power your local wallet.

bsvd does not include any wallet functionality by design as it makes the codebase more modular and easy to maintain. 
The [bsvwallet](https://github.com/metasv/bsvwallet) is a separate application that provides a secure Bitcoin (BSV) wallet 
that communicates with your running bsvd instance via the API.

## Table of Contents

- [Requirements](#requirements)
- [Install](#install)
  - [Install prebuilt packages](#install-pre-built-packages)
  - [Build from Source](#build-from-source)
- [Getting Started](#getting-started)
- [Protoconf Support](#protoconf-support)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Requirements

[Go](http://golang.org) 1.9 or newer.

## Install

### Install Pre-built Packages

The easiest way to run the server is to download a pre-built binary. You can find binaries of our latest release for each operating system at the [releases page](https://github.com/metasv/mvcd/releases).

### Build from Source

If you prefer to install from source do the following:

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Run the following commands to obtain btcd, all dependencies, and install it:

```bash
$ go get github.com/metasv/mvcd
```

This will download and compile `bsvd` and put it in your path.

If you are a bsvd contributor and would like to change the default config file (`bsvd.conf`), make any changes to `sample-bsvd.conf` and then run the following commands:

```bash
$ go-bindata sample-bsvd.conf  # requires github.com/go-bindata/go-bindata/
$ gofmt -s -w bindata.go
```

## Getting Started

To start bsvd with default options just run:

```bash
$ ./bsvd
```

You'll find a large number of runtime options on the help menu. All of which can also be set in a config file.
See the [sample config file](https://github.com/metasv/mvcd/blob/master/sample-bsvd.conf) for an example of how to use it.

## Docker

Building and running `bsvd` in docker is quite painless. To build the image:

```
docker build . -t bsvd
```

To run the image:

```
docker run bsvd
```

To run `bsvctl` and connect to your `bsvd` instance:

```
# Find the running bsvd container.
docker ps

# Exec bsvctl.
docker exec <container> bsvctl <command>
```

## Protoconf Support

In the higher version (0.2.2+) of BSV node, a new wire message "protoconf" has been added. We forked the original bitcoinsv/bsvd, and made some modification to support "protoconf" message. Now you can connect to higher version of BSV node through P2P protocol. Here are some test code

```
package main

import (
	"github.com/metasv/mvcd/chaincfg"
	"github.com/metasv/mvcd/peer"
	"github.com/metasv/mvcd/wire"
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
		UserAgentVersion: "1.0.0",
		ChainParams:      &chaincfg.MainNetParams,
		Services:         wire.SFNodeBitcoinCash|wire.SFNodeNetwork|wire.SFNodeBloom,
		TrickleInterval:  time.Second * 10,
		Listeners: peer.MessageListeners{
			OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
				log.Printf("outbound: received version")
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
			OnProtoconf: func(p *peer.Peer, msg *wire.MsgProtoconf) {
				log.Println("protoconf")
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
				log.Println("received new block of block hash " + msg.BlockHash().String())
			},
		},
	}
	p, err := peer.NewOutboundPeer(peerCfg, "your-node-ip:8333")
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
```

## Documentation

The documentation is a work-in-progress.  It is located in the [docs](https://github.com/metasv/mvcd/tree/master/docs) folder.

## Contributing

Contributions are definitely welcome! Please read the contributing [guidelines](https://github.com/metasv/mvcd/blob/master/docs/code_contribution_guidelines.md) before starting.

## Security Disclosures

To report security issues please contact:

Chris Pacia (ctpacia@gmail.com) - GPG Fingerprint: 0150 2502 DD3A 928D CE52 8CB9 B895 6DBF EE7C 105C

or

Josh Ellithorpe (quest@mac.com) - GPG Fingerprint: B6DE 3514 E07E 30BB 5F40  8D74 E49B 7E00 0022 8DDD 

## License

bsvd is licensed under the [copyfree](http://copyfree.org) ISC License.
