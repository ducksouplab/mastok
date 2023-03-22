const VERSION = "1";
const LOCAL_KEY = "mastok_random_id";

const print = (input) => {
  if (Array.isArray(input)) {
    let output = "";
    for (item of input) {
      output += print(item);
    }
    return output;
  } else if (typeof input === "object") {
    let output = "";
    for (const i in input) {
      output += print(input[i]);
    }
    return output;
  }
  // default
  return `${input}`
};

const getWebGLInfo = () => {
  let output = "";
  let gl = document.createElement("canvas").getContext("webgl");
  if(gl) {
    output += `${gl.getParameter(gl.VERSION)}${gl.getParameter(gl.VENDOR)}${gl.getParameter(gl.RENDERER)}`;
    debug = gl.getExtension('WEBGL_debug_renderer_info');
    if(debug) {
      output += gl.getParameter(debug.UNMASKED_RENDERER_WEBGL);
    }
  }
  return output;
}

const getLocalId = () => {
  let localId = localStorage.getItem(LOCAL_KEY);
  // get previous
  if(localId) {
    return localId;
  }
  // or create one
  localId = crypto.randomUUID();
  localStorage.setItem(LOCAL_KEY, localId);
  return localId;
}

const concatenateMetrics = () => {
  // initializes with script version
  let metrics = VERSION;
  // get id from localStorage
  metrics += getLocalId();
  // server sent metrics
  metrics += `${serverHash}`;
  // user agent
  metrics += navigator.userAgent;
  // navigator properties
  metrics += navigator.deviceMemory;
  metrics += navigator.hardwareConcurrency;
  metrics += navigator.language;
  metrics += navigator.maxTouchPoints;
  metrics += navigator.pdfViewerEnabled;
  // screen properties
  metrics += window.screen.width;
  metrics += window.screen.height;
  metrics += window.screen.pixelDepth;
  metrics += devicePixelRatio;
  // timezone
  metrics += (new Date()).getTimezoneOffset();
  if(Intl) {
    metrics += Intl.DateTimeFormat().resolvedOptions().timeZone
  }
  // WebGL
  metrics += getWebGLInfo();
  return metrics;
};

// from https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest#converting_a_digest_to_a_hex_string
const digest = async (message) => {
  const msgUint8 = new TextEncoder().encode(message); // encode as (utf-8) Uint8Array
  const hashBuffer = await crypto.subtle.digest("SHA-512", msgUint8); // hash the message
  const hashArray = Array.from(new Uint8Array(hashBuffer)); // convert buffer to byte array
  const hashHex = hashArray
    .map((b) => b.toString(16).padStart(2, "0"))
    .join(""); // convert bytes to hex string
  return hashHex;
}

const fingerprint = async () => {
  const metrics = concatenateMetrics();
  hash = await digest(metrics);
  return hash;
};

export default fingerprint;
