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
    const { kind, payload } = looseJSONParse(event.data)
    if(kind === 'State') {
      document.getElementById("state").innerHTML = payload;
      if(payload === "Paused") {
        document.getElementById("change-state-container").value = "Run";
        show("change-state-container");
      } else if(payload === "Running") {
        document.getElementById("change-state-container").value = "Pause";
        show("change-state-container");
      } else if(payload === "Completed") {
        hide("not-completed-container");
        hide("change-state-container");
        show("completed-container");
      }
    } else if(kind === 'PoolSize') {
      updateHasParticipants(payload);
      document.getElementById("pool-size").innerHTML = payload;
    } else if(kind === 'SessionStart') {
      hide("size-container");
      show("new-container");
      setTimeout(() => window.location.reload(), 3000);
    }
  };

  document.getElementById("change-state-container").addEventListener('click', (e) => {
    if(e.target.value == "Run") {
      ws.send(JSON.stringify("State:Running"));
    } else if(e.target.value == "Pause") {
      if(state.hasParticipants) {
        if (window.confirm("Participants in the pool will be disconnected, are you sure you want to pause campaign?")) {
          ws.send(JSON.stringify("State:Paused"));
        }
      } else {
        ws.send(JSON.stringify("State:Paused"));
      }
    }
  });

};

document.addEventListener("DOMContentLoaded", async () => {
  console.log("[supervise] version 0.1");
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
