import fingerprint from "./fingerprint";

const state = {
  starting: false
};

const looseJSONParse = (str) => {
  try {
    return JSON.parse(str);
  } catch (error) {
    console.error(error);
  }
};

const containers = [
  "consent-container",
  "grouping-container",
  "waiting-container",
  "joining-container",
  "pending-container",
  "full-container",
  "unavailable-container",
  "landing-failed-container",
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
    if (!state.starting) {
      showOnly("unavailable-container");
    }
    console.log(event);
  };

  ws.onerror = (event) => {
    showOnly("unavailable-container");
    console.log(event);
  };

  ws.onmessage = (event) => {
    const { kind, payload } = looseJSONParse(event.data);
    console.log(kind, payload);
    if (kind === "Consent") {
      document.querySelector("#consent-container p").innerHTML =
        processConcent(payload);
      hide("alert-container");
      // ease checkboxes clicking
      const lis = document.querySelectorAll("#consent-container li");
      for (const li of lis) {
        li.addEventListener("click", () => {
          const checkbox = li.querySelector("input");
          if (!checkbox.disabled) {
            value = checkbox.checked;
            checkbox.checked = !value;
          }
        });
      }
      // on submit
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
            ws.send(JSON.stringify({ kind: "Agree" }));
          } else {
            show("alert-container");
          }
        });
      // show
      showOnly("consent-container");
    } else if (kind === "Grouping") {
      const [question, ...answers] = payload.split("\n");
      const action = answers.pop(); // mutates answers
      document.getElementById("grouping-question").innerText = question;
      document.getElementById("grouping-submit").innerText = action;
      const groupingAnswers = document.getElementById("grouping-answers");
      for (let a of answers) {
        const [text, _] = a.split(":");
        const answerEl = document.createElement("div");
        answerEl.classList.add("form-check");
        answerEl.innerHTML = `<input class="form-check-input" type="radio" value="${text}" name="group" id="answer-${text}" required>
        <label class="form-check-label" for="answer-${text}">${text}</label>`;
        groupingAnswers.append(answerEl);
      }
      // show submit button
      const formChecks = document.querySelectorAll("#grouping-answers input");
      for (const c of formChecks) {
        c.addEventListener("change", () => {
          document.getElementById("grouping-submit").style.display = "";
        });
      }
      // 
      document
        .querySelector("#grouping-form")
        .addEventListener("submit", (e) => {
          e.preventDefault();
          const choice = document.querySelector('input[name="group"]:checked').value;
          ws.send(JSON.stringify({ kind: "Connect", payload: choice }));
        });
      // show
      showOnly("grouping-container");
    } else if (kind === "PoolSize" && !state.starting) {
      let sizes = document.querySelectorAll(".pool-size");
      for (let s of sizes) {
        s.innerHTML = payload;
      }
      showOnly("waiting-container");
    } else if (kind === "SessionStart") {
      state.starting = true;
      // participant is joining experiment
      showOnly("joining-container");
      setTimeout(() => {
        document.location.href = payload;
      }, 3000);
    } else if (kind === "Pending") {
      showOnly("pending-container");
    } else if (kind === "Disconnect" && payload.startsWith("Redirect")) {
      // instant redirect since participant is rejoining experiment
      const target = payload.replace("Redirect:", "");
      document.location.href = target;
    } else if (kind === "Disconnect" && payload == "Full") {
      showOnly("full-container");
      ws.close();
    } else if (kind === "State" && payload == "Unavailable") {
      showOnly("unavailable-container");
      ws.close();
    } else if (kind === "State" && payload == "LandingFailed") {
      showOnly("landing-failed-container");
      ws.close();
    } else if (kind === "Disconnect") {
      ws.close();
    }
  };
};

document.addEventListener("DOMContentLoaded", async () => {
  console.log("[supervise] version 0.3.0");
  const slugMatch = /join\/(.*)$/.exec(window.location.pathname);
  if (slugMatch) {
    const slug = slugMatch[1];
    start(slug);
  }
});
