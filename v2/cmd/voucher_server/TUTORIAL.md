# Voucher Tutorial

This tutorial should give you an overview on how to set up Voucher Server.

## Installing Voucher

First, install Voucher using the instructions in the [README](/v2/cmd/voucher_server/README.md).

## Create configuration

Create a new configuration file. You may want to refer to the [example configuration](/config/config.toml).

Make sure that you update the `ejson` or `sops` specific blocks to point to the files you will be making in the next step.

## Create secrets

(Deprecated, clair is no longer supported)
If you plan on creating attestations (rather than just running checks against your images), or if you plan on using Clair as your Vulnerability Scanner, you will need to create a secrets file to store secret values like the OpenPGP keys and/or Clair login information.

### Create ejson configuration

Note: this step is unnecessary if you are a Shopify employee and are running Voucher in Shopify's cloud platform.

First, create a public and private key pair (this is unnecessary if your platform automatically generates a key for you):

```shell
$ ejson keygen
```

You will see output similar to the following:

```
Public Key:
45960c1576c3a5caa13fea5630b56ba2c48dd67b9701bddf9f24666123122306
Private Key:
0de9401520d770fe9cd4bc985dd949a3f537d338f3954ba12c65be07c5e4637f
```

You can then create an `ejson` file using the public key included. For example:

```json
{
    "_public_key": "<public key>",
    "openpgpkeys": {},
}
```

The private key should then be stored in a file, where the filename is the public key, and the body of the file is the private key.

For example:

```shell
$ echo 0de9401520d770fe9cd4bc985dd949a3f537d338f3954ba12c65be07c5e4637f > 45960c1576c3a5caa13fea5630b56ba2c48dd67b9701bddf9f24666123122306
```

You can now decrypt your `ejson` file using:

```shell
$ ejson --keydir=<path to the directory containing the keyfile> decrypt <filename of your ejson file>
```

### Create SOPS configuration

First, establish what key you'd like SOPS to use. See [SOPS usage](https://github.com/mozilla/sops#usage) for the options supported.
Voucher works best with external key storage, such as your cloud provider's KMS. This example will use the Google [Cloud Key Management](https://cloud.google.com/security-key-management) service.

You can create a `json` file containing the secrets in plaintext. For example:

```json
{
    "openpgpkeys": {
        "diy": "-----BEGIN PGP PRIVATE KEY BLOCK-----\n..."
    }
}
```

Follow your provider's instructions to [create a new key](https://cloud.google.com/kms/docs/creating-keys). This example uses:
* Project name: `my-awesome-project`
* Keyring name: `voucher`
* Key name: `voucher-sops`

You can then encrypt the file using the KMS key, and verify the result is encrypted:
```shell
$ sops -e --gcp-kms projects/my-awesome-project/locations/global/keyRings/voucher/cryptoKeys/voucher-sops config/secrets-plaintext.production.json > config/secrets.production.json
$ cat config/secrets.production.json
{
	"openpgpkeys": {
		"diy": "ENC[AES256_GCM,data:zcNXbuPAkZMuer+sXqg4F1wAblKzm93K76c0WIU=,iv:ajzwCjEaT+sPW30LrHT+F7m7tSmJDfL5AEBfU6DU7a0=,tag:ILlBpERerNBRxC8VQ7xKSw==,type:str]"
    },
    "sops": {
		"kms": null,
		"gcp_kms": [
			{
				"resource_id": "projects/my-awesome-project/locations/global/keyRings/voucher/cryptoKeys/voucher-sops",
				"created_at": "2021-09-10T20:47:03Z",
				"enc": "CiQAxEdEknNVKrA0FQG0v1FKol9sQNKsaiQerEyXG7ueLz2vBRdeLESCTN0P9D082yxFLF4QGPHtToBrUSSQBF/xdprZIxAqSn2slzIYGBTx+sr+GNy2fEakJP8UYaQDhGjBfVVsRvMwWgFYuKpF4yg="
			}
		],
		"azure_kv": null,
		"hc_vault": null,
		"age": null,
		"lastmodified": "2021-09-10T20:47:03Z",
		"mac": "ENC[AES256_GCM,data:DzYveNlcRvhrZigynfCtL4HbHS8VuoKlaozQUZD3UHQwnEraifwAQcZanHWYqW6EWj84YMG2GmT5lGYFJzMe9KTyoDXO6IDFMORDxyGaH1RddFzsn7QyLFttwvxQ+5u/J0xpQTEzzZUAJravHtx+xg4i6W0Uv22FS15HaoFMObZQh+9tJHMjzqVduNN48VkVs=,iv:ldhb/UUqYBoZWUe/SNwo=,tag:G/VgUd/c0JwFJKA3ZmfRBg==,type:str]",
		"pgp": null,
		"unencrypted_suffix": "_unencrypted",
		"version": "3.7.1"
	}
}
```

Verify you can decrypt the file, then delete the plaintext copy:
```shell
$ export EDITOR="code -w" # optional, if you prefer vscode
$ sops config/secrets.production.json
$ rm config/secrets-plaintext.production.json
```

In the future if you need to edit the file, use the same `sops config/secrets.production.json` command. SOPS files embed a message authentication code, so you can only edit in this way.

## Generating Keys for Attestation

Attestation uses GPG signing keys to sign the image. It's suggested that you use a primary GPG key instead of a subkey.

To generate a signing key, use the following command. This will ensure that you're generating a new key that only can
be used for signing (as the other attributes are unnecessary).

```
$ gpg --full-generate-key
```

You will first be asked what type of key you want to create. Select an RSA signing key.

```
Please select what kind of key you want:
   (1) RSA and RSA (default)
   (2) DSA and Elgamal
   (3) DSA (sign only)
   (4) RSA (sign only)
Your selection? 4
```

Next you will be prompted for the key length. We'll use the largest possible value, 4096 bits.

```
RSA keys may be between 1024 and 4096 bits long.
What keysize do you want? (2048) 4096
Requested keysize is 4096 bits       
```

When prompted for how long the key should be valid, select an [appropriate cryptoperiod](https://www.keylength.com/) for your deployment. For this tutorial we'll specify that the key does not expire.

```
Please specify how long the key should be valid.
         0 = key does not expire
      <n>  = key expires in n days
      <n>w = key expires in n weeks
      <n>m = key expires in n months
      <n>y = key expires in n years
Key is valid for? (0) 0
Key does not expire at all
Is this correct? (y/N) y
```

You will next be asked to provide an ID for the GPG key. You will want to add a comment to clarify which Check this key is for.

```                        
GnuPG needs to construct a user ID to identify your key.

Real name: Cloud Security 
Email address: cloudsecurityteam@example.com
Comment: DIY                       
You selected this USER-ID:
    "Cloud Security (DIY) <cloudsecurityteam@example.com>"

Change (N)ame, (C)omment, (E)mail or (O)kay/(Q)uit? o
```

The system will generate the private key. Once that has completed, you'll get a message similar to the following:

```
Note that this key cannot be used for encryption.  You may want to use
the command "--edit-key" to generate a subkey for this purpose.
pub   rsa4096/0x2A468DCA15B582C7 2018-08-17 [SC]
      Key fingerprint = 2032 24C4 5F50 3F4E 4D2F  534D 2A46 8DCA 15B5 82C7
uid                   Cloud Security (DIY) <cloudsecurityteam@example.com>
```

You can then export that key for use in Voucher, by running:

```
$ gpg -a --export-secret-key 0x2A468DCA15B582C7 > diy.gpg
```

If you look in `diy.gpg`, you will see something similar to the following:

```
-----BEGIN PGP PRIVATE KEY BLOCK-----

lQcYBFt23t8BEADuZqi....
```

This key will need to be put into the ejson file, so you will need to replace all of the newlines with "\n".

Our example key from before would then look like this.

```
-----BEGIN PGP PRIVATE KEY BLOCK-----\nlQcYBFt23t8BEADuZqi....
```

Next, create a new value in the `openpgpkeys` block in your secrets file. Make sure the key name is the same as it's name in the source code (eg, for the "DIY" test, use "diy"):

```json
{
    "openpgpkeys": {
        "diy": "-----BEGIN PGP PRIVATE KEY BLOCK-----\nlQcYBFt23t8BEADuZqi...."
    },
}
```

For ejson, edit the file and call `ejson encrypt` to encrypt it:

```shell
$ code secrets.ejson
$ ejson encrypt secrets.ejson
```

For SOPS, call `sops` to edit the file.

```shell
$ sops secrets.json
```
