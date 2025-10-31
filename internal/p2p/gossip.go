package p2p

import (
	"context"
	"encoding/json"
	"log"

	"github.com/adlonymous/discoverx402/internal/state"
	"github.com/adlonymous/discoverx402/internal/types"
	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	peer "github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const AnnounceTopic = "/x402/announce/v1"

type Node struct {
	hostAddr  string
	repo      *state.Repo
	ps        *pubsub.PubSub
	topic     *pubsub.Topic
	sub       *pubsub.Subscription
	ownPeerID peer.ID
}

func (n *Node) ListPeers() []peer.ID {
	return n.topic.ListPeers()
}

func Start(ctx context.Context, repo *state.Repo, bootstrapMA string) (*Node, error) {
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	topic, err := ps.Join(AnnounceTopic)
	if err != nil {
		return nil, err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	ownPeerID := h.ID()
	var adv string
	if addrs := h.Addrs(); len(addrs) > 0 {
		adv = addrs[0].String() + "/p2p/" + ownPeerID.String()
		log.Printf("p2p listening: %s", adv)
	} else {
		log.Printf("p2p listening (no public address yet), peerID=%s", ownPeerID.String())
	}

	if bootstrapMA != "" {
		if m, err := ma.NewMultiaddr(bootstrapMA); err == nil {
			if err := h.Connect(ctx, *maToAddrInfo(m)); err != nil {
				log.Printf("p2p connect failed: %v", err)
			} else {
				log.Printf("p2p connected to bootstrap node: %s", bootstrapMA)
			}
		}
	}

	n := &Node{hostAddr: adv, repo: repo, ps: ps, topic: topic, sub: sub, ownPeerID: ownPeerID}

	go n.readLoop(ctx)

	return n, nil
}

func (n *Node) readLoop(ctx context.Context) {
	for {
		msg, err := n.sub.Next(ctx)
		if err != nil {
			log.Printf("p2p readLoop error: %v", err)
			return
		}

		if msg.GetFrom() == n.ownPeerID {
			log.Printf("p2p ignoring self-originated message from own peer %s", msg.GetFrom())
			continue
		}

		log.Printf("p2p received message from %s (via %s)", msg.GetFrom(), msg.ReceivedFrom)

		var l types.Listing
		if err := json.Unmarshal(msg.Data, &l); err != nil {
			log.Printf("p2p unmarshal error: %v", err)
			continue
		}

		if l.Resource == "" {
			log.Printf("p2p ignoring message with empty resource")
			continue
		}

		if err := n.repo.Upsert(ctx, l); err != nil {
			log.Printf("p2p upsert error: %v", err)
		} else {
			log.Printf("p2p upserted listing: %s", l.Resource)
		}
	}

}

func (n *Node) Publish(ctx context.Context, l types.Listing) error {
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	peers := n.topic.ListPeers()
	log.Printf("p2p publishing listing %s to %d peers in mesh", l.Resource, len(peers))
	err = n.topic.Publish(ctx, b)
	if err != nil {
		log.Printf("p2p publish error: %v", err)
	}
	return err
}

func maToAddrInfo(m ma.Multiaddr) *peer.AddrInfo {
	ai, _ := peer.AddrInfoFromP2pAddr(m)
	return ai
}
