const start = function (namespace) {
  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  // check if app is served under a prefix path (if not, pathPrefix is "/")
  const pathPrefixhMatch = /(.*)campaigns/.exec(window.location.pathname);
  const pathPrefix = pathPrefixhMatch[1];
  // connect to ws endpoint
  const wsUrl = `${wsProtocol}://${window.location.host}${pathPrefix}ws/campaigns/supervise?namespace=${namespace}`;
  const ws = new WebSocket(wsUrl);

  ws.onclose = (event) => {
    console.log(event);
  };

  ws.onerror = (event) => {
    console.log(event);
  };

  ws.onmessage = (event) => {
    console.log(event);
  };

  ws.onopen = (event) => {
    console.log(event);
  };
};

document.addEventListener("DOMContentLoaded", async () => {
  const namespaceMatch = /campaigns\/supervise\/(.*)$/.exec(window.location.pathname);
  if(namespaceMatch) {
    const namespace = namespaceMatch[1];
    start(namespace);
  }
});
