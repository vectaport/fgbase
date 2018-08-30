package fgbase

import (
	"golang.org/x/crypto/nacl/box"
)

type enc struct {
	LocalPrivateKey, RemotePublicKey *[32]byte
	Nonce                            [24]byte
}

func encryptFire(n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	s := n.Aux.(*enc)
	privateKey := s.LocalPrivateKey
	publicKey := s.RemotePublicKey
	nonce := s.Nonce

	av := a.SrcGet()
	before := []byte(av.(string))
	after := box.Seal(nil, before, &nonce, publicKey, privateKey)
	x.DstPut(after)

}

// FuncEncrypt encrypts a buffer of byte data
func FuncEncrypt(a Edge, x Edge, localPrivateKey, remotePublicKey *[32]byte, nonce [24]byte) Node {

	node := MakeNode("encrypt", []*Edge{&a}, []*Edge{&x}, nil, encryptFire)
	node.Aux = &enc{LocalPrivateKey: localPrivateKey, RemotePublicKey: remotePublicKey, Nonce: nonce}
	return node

}
