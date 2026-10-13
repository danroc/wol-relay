//! Wake-on-LAN packet building and parsing.

/// Wake-on-LAN constants.
pub const DEFAULT_PORT: u16 = 9;
pub const PACKET_SIZE: usize = 102;
pub const HEADER_SIZE: usize = 6;
pub const MAC_SIZE: usize = 6;
pub const MAC_REPEAT: usize = 16;

/// A MAC-48 address.
pub type Mac = [u8; MAC_SIZE];

/// A Wake-on-LAN packet.
pub type Packet = [u8; PACKET_SIZE];

/// Wake-on-LAN errors.
#[derive(Debug, Clone, Copy, PartialEq, Eq, thiserror::Error)]
pub enum WolError {
    #[error("invalid MAC-48 address")]
    InvalidMac,
    #[error("invalid Wake-on-LAN packet size")]
    InvalidSize,
    #[error("invalid Wake-on-LAN packet header")]
    InvalidHeader,
}

/// Formats a MAC address as colon-separated lowercase hex, e.g. "00:11:22:33:44:55".
pub fn mac_to_string(mac: &Mac) -> String {
    mac.iter()
        .map(|b| format!("{b:02x}"))
        .collect::<Vec<_>>()
        .join(":")
}

/// Builds a Wake-on-LAN packet with the given MAC address.
pub fn build_packet(mac: Mac) -> Packet {
    let mut packet = [0xffu8; PACKET_SIZE];
    for (i, byte) in packet.iter_mut().enumerate().skip(HEADER_SIZE) {
        *byte = mac[(i - HEADER_SIZE) % MAC_SIZE];
    }
    packet
}

/// Parses a Wake-on-LAN packet and returns the MAC address contained in it.
pub fn parse_packet(packet: &[u8]) -> Result<Mac, WolError> {
    if packet.len() != PACKET_SIZE {
        return Err(WolError::InvalidSize);
    }

    // The header must be the 0xFF byte repeated 6 times.
    if packet[..HEADER_SIZE] != [0xff; HEADER_SIZE] {
        return Err(WolError::InvalidHeader);
    }

    // The MAC address should be the same MAC address repeated 16 times.
    let mac: Mac = packet[HEADER_SIZE..HEADER_SIZE + MAC_SIZE]
        .try_into()
        .expect("slice length equals MAC_SIZE");
    for repeat in 1..MAC_REPEAT {
        let start = HEADER_SIZE + MAC_SIZE * repeat;
        let candidate: &Mac = packet[start..start + MAC_SIZE]
            .try_into()
            .expect("slice length equals MAC_SIZE");
        if candidate != &mac {
            return Err(WolError::InvalidMac);
        }
    }

    Ok(mac)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn build_and_parse_roundtrip() {
        let mac: Mac = [0x00, 0x11, 0x22, 0x33, 0x44, 0x55];
        let packet = build_packet(mac);
        assert_eq!(packet.len(), PACKET_SIZE);
        assert_eq!(parse_packet(&packet).unwrap(), mac);
    }

    #[test]
    fn parse_packet_rejects_wrong_size() {
        let short = [0u8; PACKET_SIZE - 1];
        assert_eq!(parse_packet(&short), Err(WolError::InvalidSize));
        let long = vec![0xffu8; PACKET_SIZE + 1];
        assert_eq!(parse_packet(&long), Err(WolError::InvalidSize));
    }

    #[test]
    fn parse_packet_rejects_bad_header() {
        let mut packet = build_packet([0x00, 0x11, 0x22, 0x33, 0x44, 0x55]);
        packet[0] = 0x00;
        assert_eq!(parse_packet(&packet), Err(WolError::InvalidHeader));
    }

    #[test]
    fn parse_packet_rejects_mismatched_repeat() {
        let mut packet = build_packet([0x00, 0x11, 0x22, 0x33, 0x44, 0x55]);
        packet[HEADER_SIZE + MAC_SIZE * 5] = 0x99;
        assert_eq!(parse_packet(&packet), Err(WolError::InvalidMac));
    }

    #[test]
    fn mac_displays_colon_separated_hex() {
        let mac: Mac = [0x00, 0x11, 0x22, 0x33, 0x44, 0x55];
        assert_eq!(mac_to_string(&mac), "00:11:22:33:44:55");
    }
}
