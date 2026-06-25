/*
Package local defines an implementation of an on-disk, EIP-2335 keystore.json
approach towards defining validator accounts in Qrysm. A validating private key is
encrypted using a passphrase and its resulting encrypted file is stored as a
keystore.json file under a unique, human-readable, account namespace. This local keymanager approach
relies on storing account information on-disk, making it trivial to import, backup and
list all associated accounts for a user.

Qrysm uses an EIP-2335-style keystore format defined by https://eips.ethereum.org/EIPS/eip-2335
for storing and defining encryption for QRL ML-DSA-87 validator private keys. Passwords
utilized for key encryptions are strings of arbitrary unicode characters.
The password is first converted to its NFKD representation, stripped of control codes specified
in the EIP link above, and finally the password is UTF-8 encoded.
*/
package local
