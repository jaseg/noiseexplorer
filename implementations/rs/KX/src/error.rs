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

#[derive(Debug)]
pub enum NoiseError {
	DecryptionError,
	UnsupportedMessageLengthError,
	ExhaustedNonceError,
	InvalidKeyError,
    InvalidInputError,
    DerivePublicKeyFromEmptyKeyError,
    Hex(hex::FromHexError),
}

impl std::fmt::Display for NoiseError {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        match *self {
            NoiseError::DecryptionError => write!(f, "Unsuccesful decryption."),
            NoiseError::UnsupportedMessageLengthError => write!(f, "Unsupported Message Length."),
            NoiseError::ExhaustedNonceError => write!(f, "Reached maximum number of messages that can be sent for this session."),
            NoiseError::DerivePublicKeyFromEmptyKeyError => write!(f, "Unable to derive PublicKey"),
            NoiseError::InvalidKeyError => write!(f, "Invalid Key."),
            NoiseError::InvalidInputError => write!(f, "Invalid input length."),
            NoiseError::Hex(ref e) => e.fmt(f),
        }
    }
}

impl std::error::Error for NoiseError {
    fn description(&self) -> &str {
        match *self {
            NoiseError::DecryptionError => "The sender was not authenticated.",
            NoiseError::UnsupportedMessageLengthError => "The length of a transport message must be exclusively between 0 and 65535 bytes.",
            NoiseError::ExhaustedNonceError => "You must end terminate this session and start a new one.",
            NoiseError::DerivePublicKeyFromEmptyKeyError => "It is forbidden to derive a PublicKey from an Empty PrivateKey",
            NoiseError::InvalidKeyError => "The key must be exactly 32 bytes of length.",
            NoiseError::InvalidInputError => "The input string exactly 32 bytes of length.",
            NoiseError::Hex(ref e) => e.description(),
        }
    }
}

impl From<hex::FromHexError> for NoiseError {
    fn from(err: hex::FromHexError) -> NoiseError {
        NoiseError::Hex(err)
    }
}
