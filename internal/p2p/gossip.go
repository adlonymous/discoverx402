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
	hostAddr string
	repo     *state.Repo
	ps       *pubsub.PubSub
	topic    *pubsub.Topic
	sub      *pubsub.Subscription
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

	var adv string
	if addrs := h.Addrs(); len(addrs) > 0 {
		adv = addrs[0].String() + "/p2p/" + h.ID().String()
		log.Printf("p2p listening: %s", adv)
	} else {
		log.Printf("p2p listening (no public address yet), peerID=%s", h.ID().String())
	}

	if bootstrapMA != "" {
		if m, err := ma.NewMultiaddr(bootstrapMA); err == nil {
			if err := h.Connect(ctx, *maToAddrInfo(m)); err != nil {
				if err := h.Connect(ctx, *maToAddrInfo(m)); err != nil {
					log.Printf("p2p connect failed: %v", err)
				} else {
					log.Printf("p2p connected to bootstrap node: %s", bootstrapMA)
				}
			}
		}
	}

	n := &Node{hostAddr: adv, repo: repo, ps: ps, topic: topic, sub: sub}

	go n.readLoop(ctx)

	return n, nil
}

func (n *Node) readLoop(ctx context.Context) {
	for {
		msg, err := n.sub.Next(ctx)
		if err != nil {
			return
		}

		if msg.ReceivedFrom == msg.GetFrom() {
			continue
		}

		var l types.Listing
		if err := json.Unmarshal(msg.Data, &l); err != nil {
			continue
		}

		if l.Resource == "" {
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
	return n.topic.Publish(ctx, b)
}

func maToAddrInfo(m ma.Multiaddr) *peer.AddrInfo {
	ai, _ := peer.AddrInfoFromP2pAddr(m)
	return ai
}
