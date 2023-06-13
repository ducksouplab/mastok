const showOtreeDoc = (e) => {
  // hide all
  const docs = document.querySelectorAll(".campaign-form .form-text");
  for (const d of docs) {
    d.style.display = "none";
  }
  // show one
  const selected = e.target.value;
  const doc = document.querySelector(`.campaign-form .form-text.${selected}`);
  doc.style.display = "block";
}

const init = () => {
  const selectOtree = document.querySelector(".campaign-form .otree-select select");
  if (selectOtree) {
    selectOtree.addEventListener("change", showOtreeDoc);
    document.querySelector(".campaign-form .otree-select .otree-doc .form-text:first-child").style.display = "block";
  }
}

document.addEventListener("DOMContentLoaded", init)
