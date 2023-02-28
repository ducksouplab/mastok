const start = function (slug) {
  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  // check if app is served under a prefix path (if not, pathPrefix is "/")
  const pathPrefixhMatch = /(.*)join/.exec(window.location.pathname);
  const pathPrefix = pathPrefixhMatch[1];
  // connect to ws endpoint
  const wsUrl = `${wsProtocol}://${window.location.host}${pathPrefix}ws/join?slug=${slug}`;
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
  const slugMatch = /join\/(.*)$/.exec(window.location.pathname);
  if(slugMatch) {
    const slug = slugMatch[1];
    start(slug);
  }
});
