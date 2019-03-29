#![allow(non_snake_case, non_upper_case_globals)]

use noiseexplorer_i1x1;
use hex;

fn decode_str(s: &str) -> Vec<u8> {
    if let Ok(x) = hex::decode(s) {
        x
    } else {
        panic!("{:X?}", hex::decode(s).err());
    }
}

fn decode_str_32(s: &str) -> [u8; 32] {
	if let Ok(x) = hex::decode(s) {
		if x.len() == 32 {
			let mut temp: [u8; 32] = [0u8; 32];
			temp.copy_from_slice(&x[..]);
			temp
		} else {
			panic!("Invalid input length; decode_32");
		}
	} else {
		panic!("Invalid input length; decode_32");
	}
}

#[test]
fn test() {
    let prologue = decode_str("4a6f686e2047616c74");
	let initStaticA: noiseexplorer_i1x1::Keypair = noiseexplorer_i1x1::Keypair::new_k(decode_str_32("e61ef9919cde45dd5f82166404bd08e38bceb5dfdfded0a34c8df7ed542214d1"));
	let respStatic: noiseexplorer_i1x1::Keypair = noiseexplorer_i1x1::Keypair::new_k(decode_str_32("4a3acbfdb163dec651dfa3194dece676d437029c62a408b4c5ea9114246e4893"));
	let mut initiatorSession: noiseexplorer_i1x1::NoiseSession =
	noiseexplorer_i1x1::NoiseSession::InitSession(true, &prologue, initStaticA, noiseexplorer_i1x1::EMPTY_KEY);
	let mut responderSession: noiseexplorer_i1x1::NoiseSession =
	noiseexplorer_i1x1::NoiseSession::InitSession(false, &prologue, respStatic, noiseexplorer_i1x1::EMPTY_KEY);
	initiatorSession.set_ephemeral_keypair(noiseexplorer_i1x1::Keypair::new_k(decode_str_32(
		"893e28b9dc6ca8d611ab664754b8ceb7bac5117349a4439a6b0569da977c464a"
	)));
	responderSession.set_ephemeral_keypair(noiseexplorer_i1x1::Keypair::new_k(decode_str_32(
		"bbdb4cdbd309f1a1f2e1456967fe288cadd6f712d65dc7b7793d5e63da6b375b"
	)));
	let payloadA = decode_str("4c756477696720766f6e204d69736573");
	let mut messageA: noiseexplorer_i1x1::MessageBuffer = initiatorSession.SendMessage(&payloadA);
	let mut validA: bool = false;
	if let Some(_x) = responderSession.RecvMessage(&mut messageA) {
		validA = true;
	}
	let tA: Vec<u8> = decode_str("ca35def5ae56cec33dc2036731ab14896bc4c75dbb07a61f879f8e3afa4c79446bc3822a2aa7f4e6981d6538692b3cdf3e6df9eea6ed269eb41d93c22757b75a4c756477696720766f6e204d69736573");
	let payloadB = decode_str("4d757272617920526f746862617264");
	let mut messageB: noiseexplorer_i1x1::MessageBuffer = responderSession.SendMessage(&payloadB);
	let mut validB: bool = false;
	if let Some(_x) = initiatorSession.RecvMessage(&mut messageB) {
		validB = true;
	}
	let tB: Vec<u8> = decode_str("95ebc60d2b1fa672c1f46a8aa265ef51bfe38e7ccb39ec5be34069f144808843ef9b7963f1561cdea23aa6d4a0d6edebacfb3fc4d7e74e4521d0a73dfa741e461b016f8942dbbfbb1b72104bccb15e5c3a8a4c85980cdf09829c770ecd1899acf33a004032b4eea8280883270c1695");
	let payloadC = decode_str("462e20412e20486179656b");
	let mut messageC: noiseexplorer_i1x1::MessageBuffer = initiatorSession.SendMessage(&payloadC);
	let mut validC: bool = false;
	if let Some(_x) = responderSession.RecvMessage(&mut messageC) {
		validC = true;
	}
	let tC: Vec<u8> = decode_str("2e95689a35d2ee9f26af0e1a4dd2873cc2afbe8ce5dc06f1177783");
	let payloadD = decode_str("4361726c204d656e676572");
	let mut messageD: noiseexplorer_i1x1::MessageBuffer = responderSession.SendMessage(&payloadD);
	let mut validD: bool = false;
	if let Some(_x) = initiatorSession.RecvMessage(&mut messageD) {
		validD = true;
	}
	let tD: Vec<u8> = decode_str("f2cb0edc7fbbcc711aef45907e85dfc0cbc17f9807c999ea362085");
	let payloadE = decode_str("4a65616e2d426170746973746520536179");
	let mut messageE: noiseexplorer_i1x1::MessageBuffer = initiatorSession.SendMessage(&payloadE);
	let mut validE: bool = false;
	if let Some(_x) = responderSession.RecvMessage(&mut messageE) {
		validE = true;
	}
	let tE: Vec<u8> = decode_str("a13afef2a6656ae8f600574b543dde2e41d1f89cc13d697042b6e2ffd29e1eeb66");
	let payloadF = decode_str("457567656e2042f6686d20766f6e2042617765726b");
	let mut messageF: noiseexplorer_i1x1::MessageBuffer = responderSession.SendMessage(&payloadF);
	let mut validF: bool = false;
	if let Some(_x) = initiatorSession.RecvMessage(&mut messageF) {
		validF = true;
	}
	let tF: Vec<u8> = decode_str("1bfedc36f5125f79910808f1b02ff8ad9297cd00739cb4cfca6c9822b8a1d554c4f8a7fe82");
	if validA && validB && validC && validD && validE && validF {
		println!("Sanity check PASS for I1X1_25519_ChaChaPoly_BLAKE2s.");
	} else {
		println!("Sanity check FAIL for I1X1_25519_ChaChaPoly_BLAKE2s.");
	}
	let mut cA: Vec<u8> = Vec::new();
	cA.append(&mut Vec::from(&messageA.ne[..]));
	cA.append(&mut messageA.ns);
	cA.append(&mut messageA.ciphertext);
	let mut cB: Vec<u8> = Vec::new();
	cB.append(&mut Vec::from(&messageB.ne[..]));
	cB.append(&mut messageB.ns);
	cB.append(&mut messageB.ciphertext);
	let mut cC: Vec<u8> = Vec::new();
	cC.append(&mut messageC.ciphertext);
	let mut cD: Vec<u8> = Vec::new();
	cD.append(&mut messageD.ciphertext);
	let mut cE: Vec<u8> = Vec::new();
	cE.append(&mut messageE.ciphertext);
	let mut cF: Vec<u8> = Vec::new();
	cF.append(&mut messageF.ciphertext);
	if tA == cA {
		println!("Test A: PASS");
	} else {
		println!("Test A: FAIL");
		println!("Expected: {:X?}", tA);
		println!("Actual:   {:X?}", cA);
	}
	if tB == cB {
		println!("Test B: PASS");
	} else {
		println!("Test B: FAIL");
		println!("Expected: {:X?}", tB);
		println!("Actual:   {:X?}", cB);
	}
	if tC == cC {
		println!("Test C: PASS");
	} else {
		println!("Test C: FAIL");
		println!("Expected: {:X?}", tC);
		println!("Actual:   {:X?}", cC);
	}
	if tD == cD {
		println!("Test D: PASS");
	} else {
		println!("Test D: FAIL");
		println!("Expected: {:X?}", tD);
		println!("Actual:   {:X?}", cD);
	}
	if tE == cE {
		println!("Test E: PASS");
	} else {
		println!("Test E: FAIL");
		println!("Expected: {:X?}", tE);
		println!("Actual:   {:X?}", cE);
	}
	if tF == cF {
		println!("Test F: PASS");
	} else {
		println!("Test F: FAIL");
		println!("Expected: {:X?}", tF);
		println!("Actual:   {:X?}", cF);
	}
	assert_eq!(tA, cA);
	assert_eq!(tB, cB);
	assert_eq!(tC, cC);
	assert_eq!(tD, cD);
	assert_eq!(tE, cE);
	assert_eq!(tF, cF);
}