/* This file is part of Noise Explorer
   Copyright 2019 Symbolic Software <symbolic.software>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

#![allow(non_snake_case, non_upper_case_globals)]

use noiseexplorer_xxpsk3::{
	noisesession::NoiseSession,
	types::{Keypair, Message, PrivateKey, PublicKey, Psk},
};

#[test]
fn noiseexplorer_test_xxpsk3() {
    if let Ok(prologue) = Message::from_str("4a6f686e2047616c74") {
	if let Ok(init_static_private) = PrivateKey::from_str("e61ef9919cde45dd5f82166404bd08e38bceb5dfdfded0a34c8df7ed542214d1") {
	if let Ok(resp_static_private) = PrivateKey::from_str("4a3acbfdb163dec651dfa3194dece676d437029c62a408b4c5ea9114246e4893") {
	if let Ok(init_static_kp) = Keypair::from_private_key(init_static_private) {
	if let Ok(resp_static_kp) = Keypair::from_private_key(resp_static_private) {
	if let Ok(psk) = Psk::from_str("54686973206973206d7920417573747269616e20706572737065637469766521") {
	
let mut initiator_session: NoiseSession = NoiseSession::init_session(true, prologue.clone(), init_static_kp, PublicKey::empty(), psk.clone());
	let mut responder_session: NoiseSession = NoiseSession::init_session(false, prologue, resp_static_kp, PublicKey::empty(), psk);
	if let Ok(initiator_ephemeral_private) = PrivateKey::from_str("893e28b9dc6ca8d611ab664754b8ceb7bac5117349a4439a6b0569da977c464a") {
if let Ok(init_ephemeral_kp) = Keypair::from_private_key(initiator_ephemeral_private) {
initiator_session.set_ephemeral_keypair(init_ephemeral_kp);
	if let Ok(responder_ephemeral_private) = PrivateKey::from_str("4a6f686e2047616c74") {
if let Ok(responder_ephemeral_kp) = Keypair::from_private_key(responder_ephemeral_private) {
responder_session.set_ephemeral_keypair(responder_ephemeral_kp);
	if let Ok(mA) = Message::from_str("4c756477696720766f6e204d69736573") {
	if let Ok(tA) = Message::from_str("ca35def5ae56cec33dc2036731ab14896bc4c75dbb07a61f879f8e3afa4c7944325ea71699951ece20f284b6ad9604a029eb335bf84564c308b6ade90ae45078") {
	if let Ok(messageA) = initiator_session.send_message(mA) {
	if let Ok(_x) = responder_session.recv_message(messageA.clone()) {
	if let Ok(mB) = Message::from_str("4d757272617920526f746862617264") {
	if let Ok(tB) = Message::from_str("95ebc60d2b1fa672c1f46a8aa265ef51bfe38e7ccb39ec5be34069f1448088432645535233ffe1432564d66a85227b677ced6fc2730ae0998ff49aa1dc56b8186e31b16e416f5d9c03c71f6c34fd37ec013105020070a8b00c000ce7ed56629c119795f96463274bc05519d5c24dc1") {
	if let Ok(messageB) = responder_session.send_message(mB) {
	if let Ok(_x) = initiator_session.recv_message(messageB.clone()) {
	if let Ok(mC) = Message::from_str("462e20412e20486179656b") {
	if let Ok(tC) = Message::from_str("adf16c5375ec4172576783fd59f2bfa5c7a320d0a13b759592e1a2ddf5524cce59ccbb92ff5d321fced3bdb2840596df562c0e68aad41b090abd285f6d300130072e06964a6ba494e58d47") {
	if let Ok(messageC) = initiator_session.send_message(mC) {
	if let Ok(_x) = responder_session.recv_message(messageC.clone()) {
	if let Ok(mD) = Message::from_str("4361726c204d656e676572") {
	if let Ok(tD) = Message::from_str("dcdc045c8e9ec36c8ea4078552e5849f87cb9bdfbd2a4eee3baaf6") {
	if let Ok(messageD) = responder_session.send_message(mD) {
	if let Ok(_x) = initiator_session.recv_message(messageD.clone()) {
	if let Ok(mE) = Message::from_str("4a65616e2d426170746973746520536179") {
	if let Ok(tE) = Message::from_str("4d11ed1f242e199dbcbc9773495834a95e8a6109e2b555aeb50780e69b152821e4") {
	if let Ok(messageE) = initiator_session.send_message(mE) {
	if let Ok(_x) = responder_session.recv_message(messageE.clone()) {
	if let Ok(mF) = Message::from_str("457567656e2042f6686d20766f6e2042617765726b") {
	if let Ok(tF) = Message::from_str("4d1e6873ffcc88490be6914928590f63253c2db434f1f206f083f89ca559a3e60a8dcc4f12") {
	if let Ok(messageF) = responder_session.send_message(mF) {
	if let Ok(_x) = initiator_session.recv_message(messageF.clone()) {
	assert!(tA == messageA, "\n\n\nTest A: FAIL\n\nExpected:\n{:X?}\n\nActual:\n{:X?}", tA, messageA);
	assert!(tB == messageB, "\n\n\nTest B: FAIL\n\nExpected:\n{:X?}\n\nActual:\n{:X?}", tB, messageB);
	assert!(tC == messageC, "\n\n\nTest C: FAIL\n\nExpected:\n{:X?}\n\nActual:\n{:X?}", tC, messageC);
	assert!(tD == messageD, "\n\n\nTest D: FAIL\n\nExpected:\n{:X?}\n\nActual:\n{:X?}", tD, messageD);
	assert!(tE == messageE, "\n\n\nTest E: FAIL\n\nExpected:\n{:X?}\n\nActual:\n{:X?}", tE, messageE);
	assert!(tF == messageF, "\n\n\nTest F: FAIL\n\nExpected:\n{:X?}\n\nActual:\n{:X?}", tF, messageF);
	}}}}}
	}}}}}}}}}}}}}}}}}}}}}}}}}}}}}
}
