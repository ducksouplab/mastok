const state = {
  hasParticipants: false
};

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

const updateHasParticipants = (payload) => {
  state.hasParticipants = !payload.startsWith("0");
}

const start = function (namespace) {
  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  // check if app is served under a prefix path (if not, pathPrefix is "/")
  const pathPrefixhMatch = /(.*)campaigns/.exec(window.location.pathname);
  const pathPrefix = pathPrefixhMatch[1];
  // connect to ws endpoint
  const wsUrl = `${wsProtocol}://${window.location.host}${pathPrefix}ws/campaigns/supervise?namespace=${namespace}`;
  const ws = new WebSocket(wsUrl);

  // ws.onopen = (event) => {};

  ws.onclose = (event) => {
    console.log(event);
  };

  ws.onerror = (event) => {
    console.log(event);
  };

  ws.onmessage = (event) => {
    const { kind, payload } = looseJSONParse(event.data);
    console.log(kind, payload);
    if(kind === "State") {
      document.getElementById("state").innerHTML = payload;
      if(payload === "Paused" || payload === "Unavailable") {
        document.getElementById("change-state-container").value = "Run";
        show("change-state-container");
        hide("size-container");
        show("paused-container");
        hide("busy-container");
      } else if(payload === "Running") {
        document.getElementById("change-state-container").value = "Pause";
        show("change-state-container");
        show("size-container");
        hide("paused-container");
        hide("busy-container");
      } else if(payload === "Busy") {
        hide("change-state-container");
        hide("size-container");
        hide("paused-container");
        show("busy-container");
      } else if(payload === "Completed") {
        hide("change-state-container");
        hide("size-container");
        hide("paused-container");
        show("completed-container");
      } 
    } else if(kind === 'JoiningSize') {
      updateHasParticipants(payload);
      document.getElementById("joining-size").innerText = payload;
    } else if(kind === 'PendingSize') {
      document.getElementById("pending-size").innerText = payload;
    } else if(kind === 'SessionStart') {
      hide("size-container");
      show("new-container");
      setTimeout(() => window.location.reload(), 3000);
    }
  };

  document.getElementById("change-state-container").addEventListener('click', (e) => {
    if(e.target.value == "Run") {
      ws.send(JSON.stringify({ kind: "State", payload: "Running"}));
    } else if(e.target.value == "Pause") {
      if(state.hasParticipants) {
        if (window.confirm("Participants in the room will be disconnected, are you sure you want to pause campaign?")) {
          ws.send(JSON.stringify({ kind: "State", payload: "Paused"}));
        }
      } else {
        ws.send(JSON.stringify({ kind: "State", payload: "Paused"}));
      }
    }
  });

};

document.addEventListener("DOMContentLoaded", async () => {
  console.log("[supervise] version 0.3.1");
  const namespaceMatch = /campaigns\/supervise\/(.*)$/.exec(
    window.location.pathname
  );
  if (namespaceMatch) {
    const namespace = namespaceMatch[1];
    start(namespace);
  }
  document.getElementById("share-url").addEventListener('click', (e) => {
    e.target.select();
  });
});
