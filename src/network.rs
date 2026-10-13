//! IPv4 network and interface resolution helpers.

use std::fmt;
use std::net::Ipv4Addr;

/// An IPv4 network with its prefix length. Host bits are masked off.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Ipv4Network {
    network: Ipv4Addr,
    prefix: u8,
}

/// Network parsing errors.
#[derive(Debug, Clone, PartialEq, Eq, thiserror::Error)]
pub enum NetworkError {
    #[error("invalid CIDR {input:?}")]
    InvalidCidr { input: String },
    #[error("IPv6 not supported: {input:?}")]
    IPv6NotSupported { input: String },
    #[error("invalid prefix length: {input:?}")]
    InvalidPrefix { input: String },
}

impl Ipv4Network {
    /// Creates a network from an address and prefix length, masking off host
    /// bits. Returns `None` if the prefix length is greater than 32.
    pub fn new(addr: Ipv4Addr, prefix: u8) -> Option<Self> {
        if prefix > 32 {
            return None;
        }
        Some(Self {
            network: Ipv4Addr::from(u32::from(addr) & mask(prefix)),
            prefix,
        })
    }

    /// The network address with host bits masked off.
    pub fn network(&self) -> Ipv4Addr {
        self.network
    }

    /// The prefix length (0 to 32).
    pub fn prefix(&self) -> u8 {
        self.prefix
    }

    /// Reports whether the given IP address is contained in this network.
    pub fn contains(&self, ip: Ipv4Addr) -> bool {
        (u32::from(ip) & mask(self.prefix)) == (u32::from(self.network) & mask(self.prefix))
    }

    /// The broadcast address of this network.
    pub fn broadcast(&self) -> Ipv4Addr {
        Ipv4Addr::from(u32::from(self.network) | !mask(self.prefix))
    }
}

impl fmt::Display for Ipv4Network {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}/{}", self.network, self.prefix)
    }
}

/// The netmask for a prefix length (0 to 32).
fn mask(prefix: u8) -> u32 {
    match prefix {
        0 => 0,
        prefix => u32::MAX << (32 - prefix),
    }
}

/// Parses a CIDR string into an IPv4 network. IPv6 CIDRs are rejected.
pub fn parse_cidr(s: &str) -> Result<Ipv4Network, NetworkError> {
    let Some((addr_part, prefix_part)) = s.split_once('/') else {
        return Err(NetworkError::InvalidCidr {
            input: s.to_string(),
        });
    };

    if addr_part.contains(':') {
        return Err(NetworkError::IPv6NotSupported {
            input: s.to_string(),
        });
    }

    let addr: Ipv4Addr = addr_part.parse().map_err(|_| NetworkError::InvalidCidr {
        input: s.to_string(),
    })?;

    let prefix: u8 = prefix_part
        .parse()
        .map_err(|_| NetworkError::InvalidPrefix {
            input: s.to_string(),
        })?;

    Ipv4Network::new(addr, prefix).ok_or(NetworkError::InvalidPrefix {
        input: s.to_string(),
    })
}

/// Converts a netmask to a prefix length. Returns `None` for non-contiguous
/// masks.
pub fn netmask_to_prefix(mask: Ipv4Addr) -> Option<u8> {
    let mask = u32::from(mask);
    let prefix = mask.leading_ones() as u8;
    if prefix == 32 || (mask << prefix) == 0 {
        Some(prefix)
    } else {
        None
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::Ipv4Addr;

    fn net(addr: &str, prefix: u8) -> Ipv4Network {
        Ipv4Network::new(addr.parse().unwrap(), prefix).unwrap()
    }

    #[test]
    fn parse_cidr_valid() {
        assert_eq!(parse_cidr("10.0.0.0/24").unwrap(), net("10.0.0.0", 24));
        assert_eq!(parse_cidr("10.0.0.5/24").unwrap(), net("10.0.0.0", 24));
        assert_eq!(
            parse_cidr("192.168.2.5/32").unwrap(),
            net("192.168.2.5", 32)
        );
        assert_eq!(parse_cidr("172.16.0.0/16").unwrap(), net("172.16.0.0", 16));
    }

    #[test]
    fn parse_cidr_masks_host_bits() {
        let parsed = parse_cidr("10.0.0.5/24").unwrap();
        assert_eq!(parsed.network(), "10.0.0.0".parse::<Ipv4Addr>().unwrap());
    }

    #[test]
    fn parse_cidr_invalid() {
        assert!(matches!(
            parse_cidr("not-a-cidr"),
            Err(NetworkError::InvalidCidr { .. })
        ));
        assert!(matches!(
            parse_cidr("::1/128"),
            Err(NetworkError::IPv6NotSupported { .. })
        ));
        assert!(matches!(
            parse_cidr("10.0.0.1/129"),
            Err(NetworkError::InvalidPrefix { .. })
        ));
        assert!(matches!(
            parse_cidr("10.0.0.1/abc"),
            Err(NetworkError::InvalidPrefix { .. })
        ));
        assert!(matches!(
            parse_cidr("10.0.0.1"),
            Err(NetworkError::InvalidCidr { .. })
        ));
    }

    #[test]
    fn display_masks_host_bits() {
        assert_eq!(
            parse_cidr("10.0.0.5/24").unwrap().to_string(),
            "10.0.0.0/24"
        );
    }

    #[test]
    fn broadcast() {
        let ip = |s: &str| s.parse::<Ipv4Addr>().unwrap();
        assert_eq!(net("192.168.1.1", 24).broadcast(), ip("192.168.1.255"));
        assert_eq!(net("10.0.0.0", 8).broadcast(), ip("10.255.255.255"));
        assert_eq!(net("192.168.2.5", 32).broadcast(), ip("192.168.2.5"));
    }

    #[test]
    fn contains() {
        let networks = [net("192.168.1.1", 24), net("10.0.0.1", 24)];
        let ip = |s: &str| s.parse::<Ipv4Addr>().unwrap();
        let cases = [
            ("192.168.1.0", true),
            ("192.168.1.1", true),
            ("192.168.1.100", true),
            ("192.168.1.255", true),
            ("192.168.0.0", false),
            ("10.0.0.0", true),
            ("10.0.0.1", true),
            ("10.0.0.100", true),
            ("10.0.0.255", true),
            ("10.0.1.0", false),
            ("172.16.0.1", false),
            ("0.0.0.0", false),
        ];
        for (s, expected) in cases {
            assert_eq!(networks.iter().any(|n| n.contains(ip(s))), expected, "{s}");
        }
    }

    #[test]
    fn netmask_to_prefix_cases() {
        let mask = |s: &str| s.parse::<Ipv4Addr>().unwrap();
        assert_eq!(netmask_to_prefix(mask("255.255.255.0")), Some(24));
        assert_eq!(netmask_to_prefix(mask("255.0.0.0")), Some(8));
        assert_eq!(netmask_to_prefix(mask("0.0.0.0")), Some(0));
        assert_eq!(netmask_to_prefix(mask("255.255.255.255")), Some(32));
        assert_eq!(netmask_to_prefix(mask("255.0.255.0")), None); // non-contiguous
    }
}
