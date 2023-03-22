import fingerprint from "./fingerprint";

const looseJSONParse = (str) => {
  try {
    return JSON.parse(str);
  } catch (error) {
    console.error(error);
  }
};

const show = (id) => {
  document.getElementById(id).style.display = '';
}

const hide = (id) => {
  document.getElementById(id).style.display = 'none';
}

const start = function (slug) {
  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  // check if app is served under a prefix path (if not, pathPrefix is "/")
  const pathPrefixhMatch = /(.*)join/.exec(window.location.pathname);
  const pathPrefix = pathPrefixhMatch[1];
  // connect to ws endpoint
  const wsUrl = `${wsProtocol}://${window.location.host}${pathPrefix}ws/join?slug=${slug}`;
  const ws = new WebSocket(wsUrl);

  ws.onopen = async () => {
    const uid = await fingerprint();
    ws.send(JSON.stringify({ kind: "Land", payload: uid }));
  };

  ws.onclose = (event) => {
    console.log(event);
  };

  ws.onerror = (event) => {
    console.log(event);
  };

  ws.onmessage = (event) => {
    const {kind, payload} = looseJSONParse(event.data);
    // console.log(kind, payload);
    if(kind === 'PoolSize') {
      document.getElementById("pool-size").innerHTML = payload;
    } else if(kind === 'SessionStart') {
      hide("size-container");
      show("connecting-container");
      setTimeout(() => {
        document.location.href = payload;
      }, 3000);
    } else if(kind === 'State' && payload == "Unavailable") {
      hide("active-container");
      show("inactive-container");
    } else if(kind === 'Participant' && payload == "Disconnect") {
      ws.close();
    }
  };
}

document.addEventListener("DOMContentLoaded", async () => {
  console.log("[supervise] version 0.2.0");
  const slugMatch = /join\/(.*)$/.exec(window.location.pathname);
  if(slugMatch) {
    const slug = slugMatch[1];
    start(slug);
  }
});
