/*
K1X1:
  -> s
  ...
  -> e
  <- e, ee, s
  -> se, es
  <-
  ->
*/

/* ---------------------------------------------------------------- *
 * PARAMETERS                                                       *
 * ---------------------------------------------------------------- */

package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"hash"
	"io"
	"math"
)

/* ---------------------------------------------------------------- *
 * TYPES                                                            *
 * ---------------------------------------------------------------- */

type keypair struct {
	pk [32]byte
	sk [32]byte
}

type messagebuffer struct {
	ne         [32]byte
	ns         []byte
	ciphertext []byte
}

type cipherstate struct {
	k [32]byte
	n uint32
}

type symmetricstate struct {
	cs cipherstate
	ck [32]byte
	h  [32]byte
}

type handshakestate struct {
	ss  symmetricstate
	s   keypair
	e   keypair
	rs  [32]byte
	re  [32]byte
	psk [32]byte
}

type noisesession struct {
	hs  handshakestate
	h   [32]byte
	cs1 cipherstate
	cs2 cipherstate
	mc  uint64
	i   bool
}

/* ---------------------------------------------------------------- *
 * CONSTANTS                                                        *
 * ---------------------------------------------------------------- */

var emptyKey = [32]byte{
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
}

var minNonce = uint32(0)

/* ---------------------------------------------------------------- *
 * UTILITY FUNCTIONS                                                *
 * ---------------------------------------------------------------- */

func getPublicKey(kp *keypair) [32]byte {
	return kp.pk
}

func isEmptyKey(k [32]byte) bool {
	return subtle.ConstantTimeCompare(k[:], emptyKey[:]) == 1
}

/* ---------------------------------------------------------------- *
 * PRIMITIVES                                                       *
 * ---------------------------------------------------------------- */

func incrementNonce(n uint32) uint32 {
	return n + 1
}

func dh(sk [32]byte, pk [32]byte) [32]byte {
	var ss [32]byte
	curve25519.ScalarMult(&ss, &sk, &pk)
	return ss
}

func generateKeypair() keypair {
	var pk [32]byte
	var sk [32]byte
	_, _ = rand.Read(sk[:])
	curve25519.ScalarBaseMult(&pk, &sk)
	return keypair{pk, sk}
}

func generatePublicKey(sk [32]byte) [32]byte {
	var pk [32]byte
	curve25519.ScalarBaseMult(&pk, &sk)
	return pk
}

func encrypt(k [32]byte, n uint32, ad []byte, plaintext []byte) []byte {
	var nonce [12]byte
	var ciphertext []byte
	enc, _ := chacha20poly1305.New(k[:])
	binary.LittleEndian.PutUint32(nonce[4:], n)
	ciphertext = enc.Seal(nil, nonce[:], plaintext, ad)
	return ciphertext
}

func decrypt(k [32]byte, n uint32, ad []byte, ciphertext []byte) (bool, []byte, []byte) {
	var nonce [12]byte
	var plaintext []byte
	enc, err := chacha20poly1305.New(k[:])
	binary.LittleEndian.PutUint32(nonce[4:], n)
	plaintext, err = enc.Open(nil, nonce[:], ciphertext, ad)
	return (err == nil), ad, plaintext
}

func getHash(a []byte, b []byte) [32]byte {
	return blake2s.Sum256(append(a, b...))
}

func hashProtocolName(protocolName []byte) [32]byte {
	var h [32]byte
	if len(protocolName) <= 32 {
		copy(h[:], protocolName)
	} else {
		h = getHash(protocolName, []byte{})
	}
	return h
}

func blake2HkdfInterface() hash.Hash {
	h, _ := blake2s.New256([]byte{})
	return h
}

func getHkdf(ck [32]byte, ikm []byte) ([32]byte, [32]byte, [32]byte) {
	var k1 [32]byte
	var k2 [32]byte
	var k3 [32]byte
	output := hkdf.New(blake2HkdfInterface, ikm[:], ck[:], []byte{})
	io.ReadFull(output, k1[:])
	io.ReadFull(output, k2[:])
	io.ReadFull(output, k3[:])
	return k1, k2, k3
}

/* ---------------------------------------------------------------- *
 * STATE MANAGEMENT                                                 *
 * ---------------------------------------------------------------- */

/* CipherState */
func initializeKey(k [32]byte) cipherstate {
	return cipherstate{k, minNonce}
}

func hasKey(cs *cipherstate) bool {
	return !isEmptyKey(cs.k)
}

func setNonce(cs *cipherstate, newNonce uint32) *cipherstate {
	cs.n = newNonce
	return cs
}

func encryptWithAd(cs *cipherstate, ad []byte, plaintext []byte) (*cipherstate, []byte) {
	e := encrypt(cs.k, cs.n, ad, plaintext)
	cs = setNonce(cs, incrementNonce(cs.n))
	return cs, e
}

func decryptWithAd(cs *cipherstate, ad []byte, ciphertext []byte) (*cipherstate, []byte, bool) {
	valid, ad, plaintext := decrypt(cs.k, cs.n, ad, ciphertext)
	cs = setNonce(cs, incrementNonce(cs.n))
	return cs, plaintext, valid
}

func reKey(cs *cipherstate) *cipherstate {
	e := encrypt(cs.k, math.MaxUint32, []byte{}, emptyKey[:])
	copy(cs.k[:], e)
	return cs
}

/* SymmetricState */

func initializeSymmetric(protocolName []byte) symmetricstate {
	h := hashProtocolName(protocolName)
	ck := h
	cs := initializeKey(emptyKey)
	return symmetricstate{cs, ck, h}
}

func mixKey(ss *symmetricstate, ikm [32]byte) *symmetricstate {
	ck, tempK, _ := getHkdf(ss.ck, ikm[:])
	ss.cs = initializeKey(tempK)
	ss.ck = ck
	return ss
}

func mixHash(ss *symmetricstate, data []byte) *symmetricstate {
	ss.h = getHash(ss.h[:], data)
	return ss
}

func mixKeyAndHash(ss *symmetricstate, ikm [32]byte) *symmetricstate {
	var tempH [32]byte
	var tempK [32]byte
	ss.ck, tempH, tempK = getHkdf(ss.ck, ikm[:])
	ss = mixHash(ss, tempH[:])
	ss.cs = initializeKey(tempK)
	return ss
}

func getHandshakeHash(ss *symmetricstate) [32]byte {
	return ss.h
}

func encryptAndHash(ss *symmetricstate, plaintext []byte) (*symmetricstate, []byte) {
	var ciphertext []byte
	if hasKey(&ss.cs) {
		_, ciphertext = encryptWithAd(&ss.cs, ss.h[:], plaintext)
	} else {
		ciphertext = plaintext
	}
	ss = mixHash(ss, ciphertext)
	return ss, ciphertext
}

func decryptAndHash(ss *symmetricstate, ciphertext []byte) (*symmetricstate, []byte, bool) {
	var plaintext []byte
	var valid bool
	if hasKey(&ss.cs) {
		_, plaintext, valid = decryptWithAd(&ss.cs, ss.h[:], ciphertext)
	} else {
		plaintext, valid = ciphertext, true
	}
	ss = mixHash(ss, ciphertext)
	return ss, plaintext, valid
}

func split(ss *symmetricstate) (cipherstate, cipherstate) {
	tempK1, tempK2, _ := getHkdf(ss.ck, []byte{})
	cs1 := initializeKey(tempK1)
	cs2 := initializeKey(tempK2)
	return cs1, cs2
}

/* HandshakeState */

func initializeInitiator(prologue []byte, s keypair, rs [32]byte, psk [32]byte) handshakestate {
	var ss symmetricstate
	var e keypair
	var re [32]byte
	name := []byte("Noise_K1X1_25519_ChaChaPoly_BLAKE2s")
	ss = initializeSymmetric(name)
	mixHash(&ss, prologue)
	mixHash(&ss, s.pk[:])
	return handshakestate{ss, s, e, rs, re, psk}
}

func initializeResponder(prologue []byte, s keypair, rs [32]byte, psk [32]byte) handshakestate {
	var ss symmetricstate
	var e keypair
	var re [32]byte
	name := []byte("Noise_K1X1_25519_ChaChaPoly_BLAKE2s")
	ss = initializeSymmetric(name)
	mixHash(&ss, prologue)
	mixHash(&ss, rs[:])
	return handshakestate{ss, s, e, rs, re, psk}
}

func writeMessageA(hs *handshakestate, payload []byte) (*handshakestate, messagebuffer) {
	ne, ns, ciphertext := emptyKey, []byte{}, []byte{}
	esk, _ := hex.DecodeString("893e28b9dc6ca8d611ab664754b8ceb7bac5117349a4439a6b0569da977c464a")
	copy(hs.e.sk[:], esk[:])
	hs.e.pk = generatePublicKey(hs.e.sk)
	ne = hs.e.pk
	mixHash(&hs.ss, ne[:])
	/* No PSK, so skipping mixKey */
	_, ciphertext = encryptAndHash(&hs.ss, payload)
	messageBuffer := messagebuffer{ne, ns, ciphertext}
	return hs, messageBuffer
}

func writeMessageB(hs *handshakestate, payload []byte) (*handshakestate, messagebuffer) {
	ne, ns, ciphertext := emptyKey, []byte{}, []byte{}
	esk, _ := hex.DecodeString("bbdb4cdbd309f1a1f2e1456967fe288cadd6f712d65dc7b7793d5e63da6b375b")
	copy(hs.e.sk[:], esk[:])
	hs.e.pk = generatePublicKey(hs.e.sk)
	ne = hs.e.pk
	mixHash(&hs.ss, ne[:])
	/* No PSK, so skipping mixKey */
	mixKey(&hs.ss, dh(hs.e.sk, hs.re))
	spk := make([]byte, len(hs.s.pk))
	copy(spk[:], hs.s.pk[:])
	_, ns = encryptAndHash(&hs.ss, spk)
	_, ciphertext = encryptAndHash(&hs.ss, payload)
	messageBuffer := messagebuffer{ne, ns, ciphertext}
	return hs, messageBuffer
}

func writeMessageC(hs *handshakestate, payload []byte) ([32]byte, messagebuffer, cipherstate, cipherstate) {
	ne, ns, ciphertext := emptyKey, []byte{}, []byte{}
	mixKey(&hs.ss, dh(hs.s.sk, hs.re))
	mixKey(&hs.ss, dh(hs.e.sk, hs.rs))
	_, ciphertext = encryptAndHash(&hs.ss, payload)
	messageBuffer := messagebuffer{ne, ns, ciphertext}
	cs1, cs2 := split(&hs.ss)
	return hs.ss.h, messageBuffer, cs1, cs2
}

func writeMessageRegular(cs *cipherstate, payload []byte) (*cipherstate, messagebuffer) {
	ne, ns, ciphertext := emptyKey, []byte{}, []byte{}
	cs, ciphertext = encryptWithAd(cs, []byte{}, payload)
	messageBuffer := messagebuffer{ne, ns, ciphertext}
	return cs, messageBuffer
}

func readMessageA(hs *handshakestate, message *messagebuffer) (*handshakestate, []byte, bool) {
	valid1 := true
	hs.re = message.ne
	mixHash(&hs.ss, hs.re[:])
	/* No PSK, so skipping mixKey */
	_, plaintext, valid2 := decryptAndHash(&hs.ss, message.ciphertext)
	return hs, plaintext, (valid1 && valid2)
}

func readMessageB(hs *handshakestate, message *messagebuffer) (*handshakestate, []byte, bool) {
	valid1 := true
	hs.re = message.ne
	mixHash(&hs.ss, hs.re[:])
	/* No PSK, so skipping mixKey */
	mixKey(&hs.ss, dh(hs.e.sk, hs.re))
	_, ns, valid1 := decryptAndHash(&hs.ss, message.ns)
	if valid1 && len(ns) == 32 {
		copy(hs.rs[:], ns)
	}
	_, plaintext, valid2 := decryptAndHash(&hs.ss, message.ciphertext)
	return hs, plaintext, (valid1 && valid2)
}

func readMessageC(hs *handshakestate, message *messagebuffer) ([32]byte, []byte, bool, cipherstate, cipherstate) {
	valid1 := true
	mixKey(&hs.ss, dh(hs.e.sk, hs.rs))
	mixKey(&hs.ss, dh(hs.s.sk, hs.re))
	_, plaintext, valid2 := decryptAndHash(&hs.ss, message.ciphertext)
	cs1, cs2 := split(&hs.ss)
	return hs.ss.h, plaintext, (valid1 && valid2), cs1, cs2
}

func readMessageRegular(cs *cipherstate, message *messagebuffer) (*cipherstate, []byte, bool) {
	/* No encrypted keys */
	_, plaintext, valid2 := decryptWithAd(cs, []byte{}, message.ciphertext)
	return cs, plaintext, valid2
}



/* ---------------------------------------------------------------- *
 * PROCESSES                                                        *
 * ---------------------------------------------------------------- */

func InitSession(initiator bool, prologue []byte, s keypair, rs [32]byte) noisesession {
	var session noisesession
	psk := emptyKey
	if initiator {
		session.hs = initializeInitiator(prologue, s, rs, psk)
	} else {
		session.hs = initializeResponder(prologue, s, rs, psk)
	}
	session.i = initiator
	session.mc = 0
	return session
}

func SendMessage(session *noisesession, message []byte) (*noisesession, messagebuffer) {
	var messageBuffer messagebuffer
	if session.mc == 0 {
		_, messageBuffer = writeMessageA(&session.hs, message)
	}
	if session.mc == 1 {
		_, messageBuffer = writeMessageB(&session.hs, message)
	}
	if session.mc == 2 {
		session.h, messageBuffer, session.cs1, session.cs2 = writeMessageC(&session.hs, message)
		session.hs = handshakestate{}
	}
	if session.mc > 2 {
		if session.i {
			_, messageBuffer = writeMessageRegular(&session.cs1, message)
		} else {
			_, messageBuffer = writeMessageRegular(&session.cs2, message)
		}
	}
	session.mc = session.mc + 1
	return session, messageBuffer
}

func RecvMessage(session *noisesession, message *messagebuffer) (*noisesession, []byte, bool) {
	var plaintext []byte
	var valid bool
	if session.mc == 0 {
		_, plaintext, valid = readMessageA(&session.hs, message)
	}
	if session.mc == 1 {
		_, plaintext, valid = readMessageB(&session.hs, message)
	}
	if session.mc == 2 {
		session.h, plaintext, valid, session.cs1, session.cs2 = readMessageC(&session.hs, message)
		session.hs = handshakestate{}
	}
	if session.mc > 2 {
		if session.i {
			_, plaintext, valid = readMessageRegular(&session.cs2, message)
		} else {
			_, plaintext, valid = readMessageRegular(&session.cs1, message)
		}
	}
	session.mc = session.mc + 1
	return session, plaintext, valid
}

func main() {
	prologue, _ := hex.DecodeString("4a6f686e2047616c74")
	var initStatic keypair
	initStaticSk, _ := hex.DecodeString("e61ef9919cde45dd5f82166404bd08e38bceb5dfdfded0a34c8df7ed542214d1")
	copy(initStatic.sk[:], initStaticSk[:])
	initStatic.pk = generatePublicKey(initStatic.sk)
	var respStatic keypair
	respStaticSk, _ := hex.DecodeString("4a3acbfdb163dec651dfa3194dece676d437029c62a408b4c5ea9114246e4893")
	copy(respStatic.sk[:], respStaticSk[:])
	respStatic.pk = generatePublicKey(respStatic.sk)
	initiatorSession := InitSession(true, prologue, initStatic, emptyKey)
	responderSession := InitSession(false, prologue, respStatic, initStatic.pk)
	payloadA, _ := hex.DecodeString("4c756477696720766f6e204d69736573")
	_, messageA := SendMessage(&initiatorSession, payloadA)
	_, _, validA := RecvMessage(&responderSession, &messageA)
	tA := "ca35def5ae56cec33dc2036731ab14896bc4c75dbb07a61f879f8e3afa4c79444c756477696720766f6e204d69736573"
	payloadB, _ := hex.DecodeString("4d757272617920526f746862617264")
	_, messageB := SendMessage(&responderSession, payloadB)
	_, _, validB := RecvMessage(&initiatorSession, &messageB)
	tB := "95ebc60d2b1fa672c1f46a8aa265ef51bfe38e7ccb39ec5be34069f14480884337f40e528241fa003e524cfb460cab5a9555687b1c4aba1637693509ce52b562f0876aad09b4f499d381c3922d11c3173a44bf355f1feac074e4bfe73f81f9450dbdf8a678db5e3d9bb88f0e159b05"
	payloadC, _ := hex.DecodeString("462e20412e20486179656b")
	_, messageC := SendMessage(&initiatorSession, payloadC)
	_, _, validC := RecvMessage(&responderSession, &messageC)
	tC := "af0a97e1d84889a64602b14c98c4de7b594f762e8aad6086153d6d"
	payloadD, _ := hex.DecodeString("4361726c204d656e676572")
	_, messageD := SendMessage(&responderSession, payloadD)
	_, _, validD := RecvMessage(&initiatorSession, &messageD)
	tD := "9022bf4f4ccb8e27a813bea4eb1b974062d12cb2644ab01bf72cc2"
	payloadE, _ := hex.DecodeString("4a65616e2d426170746973746520536179")
	_, messageE := SendMessage(&initiatorSession, payloadE)
	_, _, validE := RecvMessage(&responderSession, &messageE)
	tE := "639911393d75ca7f6bd3e4482d445fbfbe1c54291e4ea685a10f96399f777e3955"
	payloadF, _ := hex.DecodeString("457567656e2042f6686d20766f6e2042617765726b")
	_, messageF := SendMessage(&responderSession, payloadF)
	_, _, validF := RecvMessage(&initiatorSession, &messageF)
	tF := "997dc653d921d2f79df41807a15af7f257224bd5817bae1cf102a337cacef73908cf3fb12b"
	if validA && validB && validC && validD && validE && validF {
		println("Sanity check PASS for K1X1_25519_ChaChaPoly_BLAKE2s.")
	} else {
		println("Sanity check FAIL for K1X1_25519_ChaChaPoly_BLAKE2s.")
	}
	cA := hex.EncodeToString(messageA.ne[:]) + hex.EncodeToString(messageA.ns) + hex.EncodeToString(messageA.ciphertext)
	cB := hex.EncodeToString(messageB.ne[:]) + hex.EncodeToString(messageB.ns) + hex.EncodeToString(messageB.ciphertext)
	cC := hex.EncodeToString(messageC.ns) + hex.EncodeToString(messageC.ciphertext)
	cD := hex.EncodeToString(messageD.ns) + hex.EncodeToString(messageD.ciphertext)
	cE := hex.EncodeToString(messageE.ns) + hex.EncodeToString(messageE.ciphertext)
	cF := hex.EncodeToString(messageF.ns) + hex.EncodeToString(messageF.ciphertext)
	if tA == cA {
		println("Test A: PASS")
	} else {
		println("Test A: FAIL")
		println("Expected:	", tA)
		println("Actual:		", cA)
	}
	if tB == cB {
		println("Test B: PASS")
	} else {
		println("Test B: FAIL")
		println("Expected:	", tB)
		println("Actual:		", cB)
	}
	if tC == cC {
		println("Test C: PASS")
	} else {
		println("Test C: FAIL")
		println("Expected:	", tC)
		println("Actual:		", cC)
	}
	if tD == cD {
		println("Test D: PASS")
	} else {
		println("Test D: FAIL")
		println("Expected:	", tD)
		println("Actual:		", cD)
	}
	if tE == cE {
		println("Test E: PASS")
	} else {
		println("Test E: FAIL")
		println("Expected:	", tE)
		println("Actual:		", cE)
	}
	if tF == cF {
		println("Test F: PASS")
	} else {
		println("Test F: FAIL")
		println("Expected:	", tF)
		println("Actual:		", cF)
	}
}