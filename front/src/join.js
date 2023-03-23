import fingerprint from "./fingerprint";

const looseJSONParse = (str) => {
  try {
    return JSON.parse(str);
  } catch (error) {
    console.error(error);
  }
};

const containers = [
  "consent-container",
  "waiting-container",
  "joining-container",
  "unavailable-container",
];

const show = (id) => {
  document.getElementById(id).style.display = "";
};

const hide = (id) => {
  document.getElementById(id).style.display = "none";
};

const showOnly = (id) => {
  for (const c of containers) {
    c === id ? show(c) : hide(c);
  }
};

const processConcent = (html) => {
  let output = html.replaceAll("<a ", '<a target="_blank"');
  output = html.replaceAll(
    '<input type="checkbox" disabled',
    '<input type="checkbox"'
  );
  return output;
};

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
    const { kind, payload } = looseJSONParse(event.data);
    console.log(kind, payload);
    if (kind === "Consent") {
      document.querySelector("#consent-container p").innerHTML =
        processConcent(payload);
      hide("alert-container");
      showOnly("consent-container");
      // ease checkboxes clicking
      const lis = document.querySelectorAll("#consent-container li");
      for (const li of lis) {
        li.addEventListener("click", () => {
          value = li.querySelector("input").checked;
          li.querySelector("input").checked = !value;
        });
      }
      // submit
      document
        .querySelector("#consent-container button")
        .addEventListener("click", () => {
          const checkboxes = document.querySelectorAll("#consent-container input[type=\"checkbox\"]");
          let accepted = true
          for (const c of checkboxes) {
            accepted = accepted && c.checked
          }
          if(accepted) {
            hide("alert-container");
            ws.send(JSON.stringify({ kind: "Join" }));
          } else {
            show("alert-container");
          }
        });
    } else if (kind === "RoomSize") {
      let sizes = document.querySelectorAll(".room-size");
      for (let s of sizes) {
        s.innerHTML = payload;
      }
      showOnly("waiting-container");
    } else if (kind === "SessionStart") {
      showOnly("joining-container");
      setTimeout(() => {
        document.location.href = payload;
      }, 3000);
    } else if (kind === "State" && payload == "Unavailable") {
      showOnly("unavailable-container");
    } else if (kind === "Participant" && payload == "Disconnect") {
      ws.close();
    }
  };
};

document.addEventListener("DOMContentLoaded", async () => {
  console.log("[supervise] version 0.2.0");
  const slugMatch = /join\/(.*)$/.exec(window.location.pathname);
  if (slugMatch) {
    const slug = slugMatch[1];
    start(slug);
  }
});
