const parseMessage = (data) => {
  try {
    const msg = JSON.parse(data);
    const kind = msg.split(":")[0];
    const payloadStr = '"' + msg.replace(kind + ":", "") + '"';
    const payload = JSON.parse(payloadStr);
    return [kind, payload];
  } catch (error) {
    console.error(error);
  }
};

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
    const [kind, payload] = parseMessage(event.data);
    if(kind === 'State') {
      document.getElementById("state").innerHTML = payload;
      if(payload === "Paused") {
        document.getElementById("change-state").value = "Run";
        document.getElementById("change-state").style.display = ''; // show
      } else if(payload === "Running") {
        document.getElementById("change-state").value = "Pause";
        document.getElementById("change-state").style.display = ''; // show
      } else if(payload === "Completed") {
        document.getElementById("change-state").style.display = 'none'; // hide
      }
    } else if(kind === 'PoolSize') {
      document.getElementById("pool-size").innerHTML = payload;
    } else if(kind === 'SessionStart') {
      document.getElementById("show-size").style.display = 'none'; // hide
      document.getElementById("show-new").style.display = ''; // start
      setTimeout(() => window.location.reload(), 3000);
    }
  };

  document.getElementById("change-state").addEventListener('click', (e) => {
    if(e.target.value == "Run") {
      ws.send(JSON.stringify("State:Running"));
    } else if(e.target.value == "Pause") {
      ws.send(JSON.stringify("State:Paused"));
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
});
