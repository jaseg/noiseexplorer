/*
N:
  <- s
  ...
  -> e, es
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
	name := []byte("Noise_N_25519_ChaChaPoly_BLAKE2s")
	ss = initializeSymmetric(name)
	mixHash(&ss, prologue)
	mixHash(&ss, rs[:])
	return handshakestate{ss, s, e, rs, re, psk}
}

func initializeResponder(prologue []byte, s keypair, rs [32]byte, psk [32]byte) handshakestate {
	var ss symmetricstate
	var e keypair
	var re [32]byte
	name := []byte("Noise_N_25519_ChaChaPoly_BLAKE2s")
	ss = initializeSymmetric(name)
	mixHash(&ss, prologue)
	mixHash(&ss, s.pk[:])
	return handshakestate{ss, s, e, rs, re, psk}
}

func writeMessageA(hs *handshakestate, payload []byte) ([32]byte, messagebuffer, cipherstate, cipherstate) {
	ne, ns, ciphertext := emptyKey, []byte{}, []byte{}
	esk, _ := hex.DecodeString("893e28b9dc6ca8d611ab664754b8ceb7bac5117349a4439a6b0569da977c464a")
	copy(hs.e.sk[:], esk[:])
	hs.e.pk = generatePublicKey(hs.e.sk)
	ne = hs.e.pk
	mixHash(&hs.ss, ne[:])
	/* No PSK, so skipping mixKey */
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

func readMessageA(hs *handshakestate, message *messagebuffer) ([32]byte, []byte, bool, cipherstate, cipherstate) {
	valid1 := true
	hs.re = message.ne
	mixHash(&hs.ss, hs.re[:])
	/* No PSK, so skipping mixKey */
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
		session.h, messageBuffer, session.cs1, _ = writeMessageA(&session.hs, message)
		session.hs = handshakestate{}
	}
	if session.mc > 0 {
		if session.i {
			_, messageBuffer = writeMessageRegular(&session.cs1, message)
		} else {
			_, messageBuffer = writeMessageRegular(&session.cs1, message)
		}
	}
	session.mc = session.mc + 1
	return session, messageBuffer
}

func RecvMessage(session *noisesession, message *messagebuffer) (*noisesession, []byte, bool) {
	var plaintext []byte
	var valid bool
	if session.mc == 0 {
		session.h, plaintext, valid, session.cs1, _ = readMessageA(&session.hs, message)
		session.hs = handshakestate{}
	}
	if session.mc > 0 {
		if session.i {
			_, plaintext, valid = readMessageRegular(&session.cs1, message)
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
	initStaticSk := emptyKey
	copy(initStatic.sk[:], initStaticSk[:])
	initStatic.pk = generatePublicKey(initStatic.sk)
	var respStatic keypair
	respStaticSk, _ := hex.DecodeString("4a3acbfdb163dec651dfa3194dece676d437029c62a408b4c5ea9114246e4893")
	copy(respStatic.sk[:], respStaticSk[:])
	respStatic.pk = generatePublicKey(respStatic.sk)
	initiatorSession := InitSession(true, prologue, initStatic, respStatic.pk)
	responderSession := InitSession(false, prologue, respStatic, emptyKey)
	payloadA, _ := hex.DecodeString("4c756477696720766f6e204d69736573")
	_, messageA := SendMessage(&initiatorSession, payloadA)
	_, _, validA := RecvMessage(&responderSession, &messageA)
	tA := "ca35def5ae56cec33dc2036731ab14896bc4c75dbb07a61f879f8e3afa4c79441b168ed8bbe8220b52bbbde6593d109d78c299b567f6e69276efcf2659c39073"
	payloadB, _ := hex.DecodeString("4d757272617920526f746862617264")
	_, messageB := SendMessage(&responderSession, payloadB)
	_, _, validB := RecvMessage(&initiatorSession, &messageB)
	tB := "a7b5d1962001e9c4d965ea5f133941e9e6989094bcde637a582c34b954f34a"
	payloadC, _ := hex.DecodeString("462e20412e20486179656b")
	_, messageC := SendMessage(&initiatorSession, payloadC)
	_, _, validC := RecvMessage(&responderSession, &messageC)
	tC := "16ff2557d5d671abe58c88d2a31b58e3a494ab3a6498124be0ea3f"
	payloadD, _ := hex.DecodeString("4361726c204d656e676572")
	_, messageD := SendMessage(&responderSession, payloadD)
	_, _, validD := RecvMessage(&initiatorSession, &messageD)
	tD := "1a6e85b0ef71c38db2c2bf3ebef1d41dc93e26bea6899187d5633d"
	payloadE, _ := hex.DecodeString("4a65616e2d426170746973746520536179")
	_, messageE := SendMessage(&initiatorSession, payloadE)
	_, _, validE := RecvMessage(&responderSession, &messageE)
	tE := "00ad2b7d0a03a748d0aefd3accee7bbbcc0bb0ed64d685b2ee8af78997a0245e3f"
	payloadF, _ := hex.DecodeString("457567656e2042f6686d20766f6e2042617765726b")
	_, messageF := SendMessage(&responderSession, payloadF)
	_, _, validF := RecvMessage(&initiatorSession, &messageF)
	tF := "5631105c749b9550b27d7926dec0c5b83d4bf207688deccd51b50dd7fc9d5e337bba9c3177"
	if validA && validB && validC && validD && validE && validF {
		println("Sanity check PASS for N_25519_ChaChaPoly_BLAKE2s.")
	} else {
		println("Sanity check FAIL for N_25519_ChaChaPoly_BLAKE2s.")
	}
	cA := hex.EncodeToString(messageA.ne[:]) + hex.EncodeToString(messageA.ns) + hex.EncodeToString(messageA.ciphertext)
	cB := hex.EncodeToString(messageB.ns) + hex.EncodeToString(messageB.ciphertext)
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