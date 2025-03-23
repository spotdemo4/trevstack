import { decode } from 'cbor2';
import { page } from '$app/state';

interface CreateCredential extends Credential {
    response: AuthenticatorAttestationResponse
}

interface AttestationObject {
    authData: Uint8Array,
    fmt: string,
    attStmt: any
}

interface DecodedPublicKeyObject {
    [key: number]: number | Uint8Array
}

export async function createPasskey(username: string, userid: number, challenge: string) {
    const challengeBuffer = Uint8Array.from(challenge, c => c.charCodeAt(0));
    const idBuffer = Uint8Array.from(userid.toString(), c => c.charCodeAt(0));

    const credential = await navigator.credentials.create({
        publicKey: {
            challenge: challengeBuffer,
            rp: { id: page.url.hostname, name: "TrevStack" },
            user: {
                id: idBuffer,
                name: username,
                displayName: username
            },
            pubKeyCredParams: [
                {
                    type: 'public-key',
                    alg: -7
                },
                {
                    type: 'public-key',
                    alg: -257
                }
            ],
            timeout: 60000,
            attestation: 'none'
        }
    }) as CreateCredential | null;

    if (!credential) {
        throw new Error('Could not create passkey');
    }

    console.log(credential.id)
    //console.log(credential.type);

    const utf8Decoder = new TextDecoder('utf-8');
    const decodedClientData = utf8Decoder.decode(credential.response.clientDataJSON)
    const clientDataObj = JSON.parse(decodedClientData);

    console.log(clientDataObj);

    const attestationObject = new Uint8Array(credential.response.attestationObject)
    const decodedAttestationObject = decode(attestationObject) as AttestationObject;

    const { authData } = decodedAttestationObject;

    // get the length of the credential ID
    const dataView = new DataView(new ArrayBuffer(2));
    const idLenBytes = authData.slice(53, 55);
    idLenBytes.forEach((value, index) => dataView.setUint8(index, value));
    const credentialIdLength = dataView.getUint16(0);

    // get the credential ID
    const credentialId = authData.slice(55, 55 + credentialIdLength);

    // get the public key object
    const publicKeyBytes = authData.slice(55 + credentialIdLength);

    console.log(publicKeyBytes);

    // the publicKeyBytes are encoded again as CBOR
    const publicKeyObject = new Uint8Array(publicKeyBytes.buffer)
    const decodedPublicKeyObject = decode(publicKeyObject) as DecodedPublicKeyObject;

    console.log(decodedPublicKeyObject);

    return {
        id: credential.id,
        publicKey: publicKeyBytes,
        algorithm: decodedPublicKeyObject[3]
    }
}

interface GetCredential extends Credential {
    response: AuthenticatorAssertionResponse
}

export async function getPasskey(passkeyids: string[], challenge: string) {
    const challengeBuffer = Uint8Array.from(challenge, c => c.charCodeAt(0));

    const credential = await navigator.credentials.get({
        publicKey: {
            challenge: challengeBuffer,
            allowCredentials: passkeyids.map((passkeyid) => {
                return {
                    id: Uint8Array.from(passkeyid, c => c.charCodeAt(0)),
                    type: 'public-key',
                }
            }),
            timeout: 60000,
        }
    }) as GetCredential | null;

    if (!credential) {
        throw new Error('Could not get passkey');
    }

    const signature = credential.response.signature;

    return {
        signature
    }
}