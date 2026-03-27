package web

var indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Chess</title>
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap');
@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&display=swap');

*{margin:0;padding:0;box-sizing:border-box;}

:root{
  --bg:#0e0e18;
  --bg2:#13131f;
  --card:#181828;
  --card-border:rgba(255,255,255,0.06);
  --hover:#1e1e38;
  --accent:#7b9aff;
  --accent2:#a78bfa;
  --accent-glow:rgba(123,154,255,0.18);
  --text:#e6e8f2;
  --text2:#8890b0;
  --text3:#505478;
  --light-sq:#f0d9b5;
  --dark-sq:#b58863;
  --light-sq-hl:#f7ec8a;
  --dark-sq-hl:#dab530;
  --check-glow:rgba(255,30,30,0.65);
  --dot-color:rgba(0,0,0,0.2);
  --ring-color:rgba(0,0,0,0.2);
  --eval-w:#eaeaea;
  --eval-b:#1c1c2e;
  --green:#34d399;
  --red:#f87171;
  --yellow:#fbbf24;
}

html,body{height:100%;}
body{
  font-family:'Inter',system-ui,-apple-system,sans-serif;
  background:var(--bg);
  color:var(--text);
  overflow-x:hidden;
}

.app{max-width:1140px;margin:0 auto;padding:14px 16px;}
.main{display:flex;gap:16px;justify-content:center;align-items:flex-start;flex-wrap:wrap;}

.eval-col{display:flex;flex-direction:column;align-items:center;gap:4px;flex-shrink:0;}
.eval-bar{
  width:26px;height:560px;background:var(--eval-b);
  border-radius:5px;overflow:hidden;position:relative;
  border:1px solid var(--card-border);
}
.eval-fill{
  position:absolute;bottom:0;left:0;right:0;height:50%;
  background:var(--eval-w);
  transition:height 0.6s cubic-bezier(.25,.8,.25,1);
}
.eval-num{
  font-family:'JetBrains Mono',monospace;font-size:0.7rem;font-weight:600;
  color:var(--text2);text-align:center;width:40px;
}

.board-col{flex-shrink:0;}

.player-bar{
  display:flex;align-items:center;gap:10px;
  padding:5px 10px;border-radius:8px;
  background:var(--card);border:1px solid var(--card-border);
  margin-bottom:3px;min-height:40px;
}
.player-bar.bottom{margin-bottom:0;margin-top:3px;}
.p-color{
  width:22px;height:22px;border-radius:4px;flex-shrink:0;
  border:1.5px solid rgba(255,255,255,0.12);
}
.p-color.white-c{background:#e8e8e8;}
.p-color.black-c{background:#2a2a2a;}
.p-name{font-size:0.8rem;font-weight:600;color:var(--text);flex:1;}
.p-captures{display:flex;gap:0;flex-wrap:wrap;max-width:170px;align-items:center;}
.p-captures svg{width:14px;height:14px;opacity:0.75;}
.p-adv{
  font-family:'JetBrains Mono',monospace;font-size:0.7rem;font-weight:600;
  color:var(--text2);margin-left:2px;
}

.board-wrap{position:relative;padding:0 0 16px 16px;}
.coord-r{
  position:absolute;left:0;width:14px;text-align:center;
  font-size:9px;font-weight:700;color:var(--text3);
  display:flex;align-items:center;justify-content:center;
  font-family:'JetBrains Mono',monospace;
}
.coord-f{
  position:absolute;bottom:0;height:14px;text-align:center;
  font-size:9px;font-weight:700;color:var(--text3);
  display:flex;align-items:center;justify-content:center;
  font-family:'JetBrains Mono',monospace;
}

.board{
  display:grid;grid-template-columns:repeat(8,1fr);grid-template-rows:repeat(8,1fr);
  width:560px;height:560px;
  border-radius:3px;overflow:hidden;
  box-shadow:0 4px 20px rgba(0,0,0,0.5);
  user-select:none;-webkit-user-select:none;position:relative;
}
.sq{
  width:70px;height:70px;position:relative;
  display:flex;align-items:center;justify-content:center;
  cursor:pointer;
}
.sq.light{background:var(--light-sq);}
.sq.dark{background:var(--dark-sq);}
.sq.last-move.light{background:var(--light-sq-hl);}
.sq.last-move.dark{background:var(--dark-sq-hl);}
.sq.selected{box-shadow:inset 0 0 0 3px var(--accent);z-index:2;}
.sq.in-check{animation:checkPulse 1s ease-in-out infinite;}
.sq.drag-over{box-shadow:inset 0 0 0 3px var(--accent);z-index:2;}

@keyframes checkPulse{
  0%,100%{box-shadow:inset 0 0 10px var(--check-glow);}
  50%{box-shadow:inset 0 0 22px rgba(255,40,40,0.85);}
}

.dot{
  position:absolute;width:17px;height:17px;border-radius:50%;
  background:var(--dot-color);pointer-events:none;z-index:5;
  animation:dotPop 0.1s ease-out;
}
.ring{
  position:absolute;inset:4px;border-radius:50%;
  border:4px solid var(--ring-color);pointer-events:none;z-index:5;
  animation:dotPop 0.1s ease-out;
}
@keyframes dotPop{from{transform:scale(0);}to{transform:scale(1);}}

.pc{
  width:56px;height:56px;z-index:10;pointer-events:auto;cursor:grab;
  filter:drop-shadow(0 1px 2px rgba(0,0,0,0.3));
  transition:none;
  will-change:transform;
}
.pc.ghost{opacity:0.3;pointer-events:none;}
.pc-anim{
  position:absolute;z-index:200;pointer-events:none;
  width:56px;height:56px;
  filter:drop-shadow(0 2px 4px rgba(0,0,0,0.35));
  will-change:transform;
}
.pc:active{cursor:grabbing;}

.panel{width:250px;display:flex;flex-direction:column;gap:8px;flex-shrink:0;}
.card{
  background:var(--card);border:1px solid var(--card-border);
  border-radius:10px;padding:12px;
}
.card-t{
  font-size:0.62rem;font-weight:700;color:var(--text3);
  text-transform:uppercase;letter-spacing:1.2px;margin-bottom:7px;
}

.status-main{font-size:0.92rem;font-weight:600;margin-bottom:3px;}
.status-sub{font-size:0.76rem;color:var(--text2);}
.status-sub.over{color:var(--yellow);font-weight:600;}

.think{display:none;align-items:center;gap:6px;padding:5px 0;color:var(--accent);font-size:0.78rem;font-weight:500;}
.think.on{display:flex;}
.dots{display:flex;gap:3px;}
.dots span{
  width:4px;height:4px;background:var(--accent);border-radius:50%;
  animation:dotBounce 1.2s infinite ease-in-out;
}
.dots span:nth-child(2){animation-delay:0.15s;}
.dots span:nth-child(3){animation-delay:0.3s;}
@keyframes dotBounce{0%,80%,100%{transform:scale(0.4);opacity:0.25;}40%{transform:scale(1);opacity:1;}}

.stats{display:grid;grid-template-columns:1fr 1fr;gap:5px;}
.st{background:rgba(255,255,255,0.02);border-radius:5px;padding:6px 8px;}
.st-l{font-size:0.58rem;color:var(--text3);text-transform:uppercase;letter-spacing:0.5px;}
.st-v{font-size:0.88rem;font-weight:600;color:var(--accent);margin-top:1px;font-family:'JetBrains Mono',monospace;}

.moves-wrap{max-height:200px;overflow-y:auto;scrollbar-width:thin;scrollbar-color:var(--hover) transparent;}
.moves-wrap::-webkit-scrollbar{width:3px;}
.moves-wrap::-webkit-scrollbar-thumb{background:var(--hover);border-radius:3px;}
.mv{
  display:flex;align-items:center;padding:2px 4px;border-radius:3px;
  font-size:0.76rem;font-family:'JetBrains Mono',monospace;
}
.mv:nth-child(odd){background:rgba(255,255,255,0.015);}
.mv-n{color:var(--text3);width:24px;flex-shrink:0;font-size:0.68rem;}
.mv-w,.mv-b{width:50px;padding:1px 3px;border-radius:2px;color:var(--text2);}
.mv-w.latest,.mv-b.latest{color:var(--accent);font-weight:600;}

.btns{display:flex;gap:5px;flex-wrap:wrap;}
.btn{
  flex:1;min-width:68px;padding:8px 8px;
  border:1px solid var(--card-border);border-radius:6px;
  background:var(--hover);color:var(--text);font-family:inherit;
  font-size:0.76rem;font-weight:500;cursor:pointer;
  transition:all 0.12s;text-align:center;
}
.btn:hover{background:var(--accent-glow);border-color:var(--accent);}
.btn:active{transform:scale(0.97);}
.btn.pri{background:var(--accent);border-color:var(--accent);color:#fff;font-weight:600;}
.btn.pri:hover{background:#93b0ff;}

.modal-bg{
  display:none;position:fixed;inset:0;background:rgba(0,0,0,0.6);
  backdrop-filter:blur(3px);z-index:1000;align-items:center;justify-content:center;
}
.modal-bg.on{display:flex;}
.modal{
  background:var(--card);border:1px solid var(--card-border);
  border-radius:12px;padding:22px;width:300px;
  box-shadow:0 14px 40px rgba(0,0,0,0.5);
  animation:modalPop 0.18s ease-out;
}
@keyframes modalPop{from{transform:scale(0.95) translateY(6px);opacity:0;}to{transform:none;opacity:1;}}
.modal h2{font-size:1.05rem;margin-bottom:14px;text-align:center;font-weight:700;}
.m-field{margin-bottom:12px;}
.m-field label{display:block;font-size:0.72rem;font-weight:500;color:var(--text2);margin-bottom:4px;}
.m-field select,.m-field input[type=range]{
  width:100%;padding:7px 9px;background:var(--bg2);
  border:1px solid var(--card-border);border-radius:6px;
  color:var(--text);font-family:inherit;font-size:0.82rem;outline:none;
}
.m-field select:focus{border-color:var(--accent);}
.m-field .range-val{text-align:center;color:var(--text2);font-size:0.8rem;margin-top:3px;font-family:'JetBrains Mono',monospace;}
.m-actions{display:flex;gap:7px;margin-top:16px;}

.promo{
  display:none;position:absolute;z-index:500;
  background:var(--card);border:1px solid rgba(255,255,255,0.12);
  border-radius:8px;padding:4px;
  box-shadow:0 8px 24px rgba(0,0,0,0.6);
  animation:modalPop 0.12s ease-out;
}
.promo.on{display:flex;flex-direction:column;gap:2px;}
.promo-item{
  width:54px;height:54px;display:flex;align-items:center;justify-content:center;
  cursor:pointer;border-radius:5px;transition:background 0.1s;
}
.promo-item:hover{background:var(--hover);}
.promo-item svg{width:44px;height:44px;}

.toast{
  position:fixed;bottom:20px;left:50%;transform:translateX(-50%) translateY(70px);
  background:var(--card);border:1px solid var(--card-border);
  border-radius:8px;padding:9px 18px;font-size:0.8rem;font-weight:500;
  box-shadow:0 6px 20px rgba(0,0,0,0.4);z-index:2000;
  opacity:0;transition:all 0.25s cubic-bezier(.25,.8,.25,1);pointer-events:none;
}
.toast.show{transform:translateX(-50%) translateY(0);opacity:1;}

@media(max-width:900px){
  .main{flex-direction:column;align-items:center;}
  .panel{width:100%;max-width:560px;}
  .eval-col{display:none;}
  .board{width:min(90vw,560px);height:min(90vw,560px);}
  .sq{width:calc(min(90vw,560px)/8);height:calc(min(90vw,560px)/8);}
  .pc{width:calc(min(90vw,560px)/8 - 14px);height:calc(min(90vw,560px)/8 - 14px);}
}
</style>
</head>
<body>
<div class="app">
  <div class="main">
    <div class="eval-col">
      <div class="eval-num" id="evalTop"></div>
      <div class="eval-bar"><div class="eval-fill" id="evalFill"></div></div>
      <div class="eval-num" id="evalBot">0.0</div>
    </div>
    <div class="board-col">
      <div class="player-bar" id="topBar">
        <div class="p-color black-c" id="topColor"></div>
        <div class="p-name" id="topName">Engine</div>
        <div class="p-captures" id="topCap"></div>
        <div class="p-adv" id="topAdv"></div>
      </div>
      <div class="board-wrap" id="boardWrap">
        <div id="coordR"></div>
        <div id="coordF"></div>
        <div class="board" id="board"></div>
        <div class="promo" id="promo"></div>
      </div>
      <div class="player-bar bottom" id="botBar">
        <div class="p-color white-c" id="botColor"></div>
        <div class="p-name" id="botName">You (White)</div>
        <div class="p-captures" id="botCap"></div>
        <div class="p-adv" id="botAdv"></div>
      </div>
    </div>
    <div class="panel">
      <div class="card">
        <div class="card-t">Game</div>
        <div class="status-main" id="sMain">White to move</div>
        <div class="think" id="think"><div class="dots"><span></span><span></span><span></span></div>Thinking&#8230;</div>
        <div class="status-sub" id="sSub"></div>
      </div>
      <div class="card">
        <div class="card-t">Engine</div>
        <div class="stats">
          <div class="st"><div class="st-l">Eval</div><div class="st-v" id="iEval">0.00</div></div>
          <div class="st"><div class="st-l">Depth</div><div class="st-v" id="iDep">&#8212;</div></div>
          <div class="st"><div class="st-l">Nodes</div><div class="st-v" id="iNod">&#8212;</div></div>
          <div class="st"><div class="st-l">NPS</div><div class="st-v" id="iNps">&#8212;</div></div>
        </div>
      </div>
      <div class="card">
        <div class="card-t">Moves</div>
        <div class="moves-wrap" id="mList"></div>
      </div>
      <div class="card">
        <div class="btns">
          <button class="btn pri" onclick="showModal()">New</button>
          <button class="btn" onclick="undo()">&#8617; Undo</button>
          <button class="btn" onclick="flip()">&#x21C5; Flip</button>
        </div>
      </div>
    </div>
  </div>
</div>
<div class="modal-bg" id="modalBg">
  <div class="modal">
    <h2>New Game</h2>
    <div class="m-field">
      <label>Play as</label>
      <select id="mColor">
        <option value="white">White</option>
        <option value="black">Black</option>
      </select>
    </div>
    <div class="m-field">
      <label>Engine Depth</label>
      <input type="range" id="mDepth" min="1" max="20" value="10" oninput="document.getElementById('mDepVal').textContent=this.value">
      <div class="range-val">Depth: <span id="mDepVal">10</span></div>
    </div>
    <div class="m-actions">
      <button class="btn" onclick="hideModal()">Cancel</button>
      <button class="btn pri" onclick="newGame()">Start</button>
    </div>
  </div>
</div>
<div class="toast" id="toast"></div>
<script>
var SVG={};
(function(){
  function mk(inner){return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 45 45">'+inner+'</svg>';}

  SVG.wK=mk('<g fill="none" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<path d="M22.5 11.63V6" stroke-linejoin="miter"/>'
    +'<path d="M20 8h5" stroke-linejoin="miter"/>'
    +'<path d="M22.5 25s4.5-7.5 3-10.5c0 0-1-2.5-3-2.5s-3 2.5-3 2.5c-1.5 3 3 10.5 3 10.5" fill="#fff" stroke-linecap="butt" stroke-linejoin="miter"/>'
    +'<path d="M12.5 37c5.5 3.5 14.5 3.5 20 0v-7s9-4.5 6-10.5c-4-6.5-13.5-3.5-16 4V27v-3.5c-2.5-7.5-12-10.5-16-4-3 6 6 10.5 6 10.5v7" fill="#fff"/>'
    +'<path d="M12.5 30c5.5-3 14.5-3 20 0M12.5 33.5c5.5-3 14.5-3 20 0M12.5 37c5.5-3 14.5-3 20 0"/>'
    +'</g>');

  SVG.bK=mk('<g fill="none" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<path d="M22.5 11.63V6" stroke-linejoin="miter"/>'
    +'<path d="M20 8h5" stroke-linejoin="miter"/>'
    +'<path d="M22.5 25s4.5-7.5 3-10.5c0 0-1-2.5-3-2.5s-3 2.5-3 2.5c-1.5 3 3 10.5 3 10.5" fill="#000" stroke-linecap="butt" stroke-linejoin="miter"/>'
    +'<path d="M12.5 37c5.5 3.5 14.5 3.5 20 0v-7s9-4.5 6-10.5c-4-6.5-13.5-3.5-16 4V27v-3.5c-2.5-7.5-12-10.5-16-4-3 6 6 10.5 6 10.5v7" fill="#000"/>'
    +'<path d="M12.5 30c5.5-3 14.5-3 20 0" stroke="#fff"/>'
    +'<path d="M12.5 33.5c5.5-3 14.5-3 20 0" stroke="#fff"/>'
    +'<path d="M12.5 37c5.5-3 14.5-3 20 0" stroke="#fff"/>'
    +'</g>');

  SVG.wQ=mk('<g fill="#fff" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<path d="M9 26c8.5-1.5 21-1.5 27 0l2-12-7 11V10l-5.5 13.5-3-15-3 15L14 10v15L7 14l2 12z" stroke-linecap="butt"/>'
    +'<path d="M9 26c0 2 1.5 2 2.5 4 1 1.5 1 1 .5 3.5-1.5 1-1.5 2.5-1.5 2.5-1.5 1.5.5 2.5.5 2.5 6.5 1 16.5 1 23 0 0 0 1.5-1 0-2.5 0 0 .5-1.5-1-2.5-.5-2.5-.5-2 .5-3.5 1-2 2.5-2 2.5-4-8.5-1.5-18.5-1.5-27 0z" stroke-linecap="butt"/>'
    +'<path d="M11.5 30c3.5-1 18.5-1 22 0M12 33.5c6-1 15-1 21 0" fill="none"/>'
    +'<circle cx="6" cy="12" r="2"/><circle cx="14" cy="9" r="2"/>'
    +'<circle cx="22.5" cy="8" r="2"/><circle cx="31" cy="9" r="2"/>'
    +'<circle cx="39" cy="12" r="2"/>'
    +'</g>');

  SVG.bQ=mk('<g fill="#000" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<circle cx="6" cy="12" r="2.75"/><circle cx="14" cy="9" r="2.75"/>'
    +'<circle cx="22.5" cy="8" r="2.75"/><circle cx="31" cy="9" r="2.75"/>'
    +'<circle cx="39" cy="12" r="2.75"/>'
    +'<path d="M9 26c8.5-1.5 21-1.5 27 0l2.5-12.5L31 25l-.3-14.1-5.2 13.6-3-14.5-3 14.5-5.2-13.6L14 25 6.5 13.5 9 26z" stroke-linecap="butt"/>'
    +'<path d="M9 26c0 2 1.5 2 2.5 4 1 1.5 1 1 .5 3.5-1.5 1-1.5 2.5-1.5 2.5-1.5 1.5.5 2.5.5 2.5 6.5 1 16.5 1 23 0 0 0 1.5-1 0-2.5 0 0 .5-1.5-1-2.5-.5-2.5-.5-2 .5-3.5 1-2 2.5-2 2.5-4-8.5-1.5-18.5-1.5-27 0z" stroke-linecap="butt"/>'
    +'<path d="M11.5 30c3.5-1 18.5-1 22 0" fill="none" stroke="#fff"/>'
    +'<path d="M12 33.5c6-1 15-1 21 0" fill="none" stroke="#fff"/>'
    +'</g>');

  SVG.wR=mk('<g fill="#fff" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<path d="M9 39h27v-3H9v3z" stroke-linecap="butt"/>'
    +'<path d="M12 36v-4h21v4H12z" stroke-linecap="butt"/>'
    +'<path d="M11 14V9h4v2h5V9h5v2h5V9h4v5" stroke-linecap="butt"/>'
    +'<path d="M34 14l-3 3H14l-3-3"/>'
    +'<path d="M15 17v7h15v-7" stroke-linecap="butt" stroke-linejoin="miter"/>'
    +'<path d="M14 29.5v-13h17v13H14z" stroke-linecap="butt" stroke-linejoin="miter"/>'
    +'<path d="M14 29.5H31M14 25.5H31M14 21.5H31" fill="none" stroke-linejoin="miter"/>'
    +'</g>');

  SVG.bR=mk('<g fill="#000" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<path d="M9 39h27v-3H9v3z" stroke-linecap="butt"/>'
    +'<path d="M12.5 32l1.5-2.5h17l1.5 2.5h-20z" stroke-linecap="butt"/>'
    +'<path d="M12 36v-4h21v4H12z" stroke-linecap="butt"/>'
    +'<path d="M14 29.5v-13h17v13H14z" stroke-linecap="butt" stroke-linejoin="miter"/>'
    +'<path d="M14 16.5L11 14h23l-3 2.5H14z" stroke-linecap="butt"/>'
    +'<path d="M11 14V9h4v2h5V9h5v2h5V9h4v5H11z" stroke-linecap="butt"/>'
    +'<path d="M12 35.5h21M13 31.5h19M14 29.5h17M14 16.5h17M11 14h23" fill="none" stroke="#fff" stroke-width="1" stroke-linejoin="miter"/>'
    +'</g>');

  SVG.wB=mk('<g fill="none" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<g fill="#fff" stroke-linecap="butt">'
    +'<path d="M9 36c3.39-.97 10.11.43 13.5-2 3.39 2.43 10.11 1.03 13.5 2 0 0 1.65.54 3 2-.68.97-1.65.99-3 .5-3.39-.97-10.11.46-13.5-1-3.39 1.46-10.11.03-13.5 1-1.354.49-2.323.47-3-.5 1.354-1.94 3-2 3-2z"/>'
    +'<path d="M15 32c2.5 2.5 12.5 2.5 15 0 .5-1.5 0-2 0-2 0-2.5-2.5-4-2.5-4 5.5-1.5 6-11.5-5-15.5-11 4-10.5 14-5 15.5 0 0-2.5 1.5-2.5 4 0 0-.5.5 0 2z"/>'
    +'<path d="M25 8a2.5 2.5 0 11-5 0 2.5 2.5 0 115 0z"/>'
    +'</g>'
    +'<path d="M17.5 26h10M15 30h15" stroke-linejoin="miter"/>'
    +'</g>');

  SVG.bB=mk('<g fill="none" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<g fill="#000" stroke-linecap="butt">'
    +'<path d="M9 36c3.39-.97 10.11.43 13.5-2 3.39 2.43 10.11 1.03 13.5 2 0 0 1.65.54 3 2-.68.97-1.65.99-3 .5-3.39-.97-10.11.46-13.5-1-3.39 1.46-10.11.03-13.5 1-1.354.49-2.323.47-3-.5 1.354-1.94 3-2 3-2z"/>'
    +'<path d="M15 32c2.5 2.5 12.5 2.5 15 0 .5-1.5 0-2 0-2 0-2.5-2.5-4-2.5-4 5.5-1.5 6-11.5-5-15.5-11 4-10.5 14-5 15.5 0 0-2.5 1.5-2.5 4 0 0-.5.5 0 2z"/>'
    +'<path d="M25 8a2.5 2.5 0 11-5 0 2.5 2.5 0 115 0z"/>'
    +'</g>'
    +'<path d="M17.5 26h10M15 30h15M22.5 15.5v5M20 18h5" stroke="#fff" stroke-linejoin="miter"/>'
    +'</g>');

  SVG.wN=mk('<g fill="none" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<path d="M22 10c10.5 1 16.5 8 16 29H15c0-9 10-6.5 8-21" fill="#fff"/>'
    +'<path d="M24 18c.38 2.91-5.55 7.37-8 9-3 2-2.82 4.34-5 4-1.042-.94 1.41-3.04 0-3-1 0 .19 1.23-1 2-1 0-4.003 1-4-4 0-2 6-12 6-12s1.89-1.9 2-3.5c-.73-.994-.5-2-.5-3 1-1 3 2.5 3 2.5h2s.78-1.992 2.5-3c1 0 1 3 1 3" fill="#fff"/>'
    +'<path d="M9.5 25.5a.5.5 0 11-1 0 .5.5 0 111 0z" fill="#000"/>'
    +'<path d="M14.933 15.75a.5 1.5 30 11-.866-.5.5 1.5 30 11.866.5z" fill="#000"/>'
    +'</g>');

  SVG.bN=mk('<g fill="none" fill-rule="evenodd" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">'
    +'<path d="M22 10c10.5 1 16.5 8 16 29H15c0-9 10-6.5 8-21" fill="#000"/>'
    +'<path d="M24 18c.38 2.91-5.55 7.37-8 9-3 2-2.82 4.34-5 4-1.042-.94 1.41-3.04 0-3-1 0 .19 1.23-1 2-1 0-4.003 1-4-4 0-2 6-12 6-12s1.89-1.9 2-3.5c-.73-.994-.5-2-.5-3 1-1 3 2.5 3 2.5h2s.78-1.992 2.5-3c1 0 1 3 1 3" fill="#000"/>'
    +'<path d="M9.5 25.5a.5.5 0 11-1 0 .5.5 0 111 0z" fill="#fff" stroke="#fff"/>'
    +'<path d="M14.933 15.75a.5 1.5 30 11-.866-.5.5 1.5 30 11.866.5z" fill="#fff" stroke="#fff"/>'
    +'<path d="M24.55 10.4l-.45 1.45.5.15c3.15 1 5.65 2.49 7.9 6.75S35.75 29.06 35.25 39l-.05.5h2.25l.05-.5c.5-10.06-.88-16.85-3.25-21.34-2.37-4.49-5.79-6.64-9.19-7.16l-.51-.1z" fill="#fff" stroke="none"/>'
    +'</g>');

  SVG.wP=mk('<path d="M22.5 9c-2.21 0-4 1.79-4 4 0 .89.29 1.71.78 2.38C17.33 16.5 16 18.59 16 21c0 2.03.94 3.84 2.41 5.03C15.41 27.09 11 31.58 11 39.5h23C34 31.58 29.59 27.09 26.59 26.03 28.06 24.84 29 23.03 29 21c0-2.41-1.33-4.5-3.28-5.62.49-.67.78-1.49.78-2.38 0-2.21-1.79-4-4-4z" fill="#fff" stroke="#000" stroke-width="1.5" stroke-linecap="round"/>');

  SVG.bP=mk('<path d="M22.5 9c-2.21 0-4 1.79-4 4 0 .89.29 1.71.78 2.38C17.33 16.5 16 18.59 16 21c0 2.03.94 3.84 2.41 5.03C15.41 27.09 11 31.58 11 39.5h23C34 31.58 29.59 27.09 26.59 26.03 28.06 24.84 29 23.03 29 21c0-2.41-1.33-4.5-3.28-5.62.49-.67.78-1.49.78-2.38 0-2.21-1.79-4-4-4z" fill="#000" stroke="#000" stroke-width="1.5" stroke-linecap="round"/>');
})();

var P_SVG={'P':'wP','N':'wN','B':'wB','R':'wR','Q':'wQ','K':'wK','p':'bP','n':'bN','b':'bB','r':'bR','q':'bQ','k':'bK'};
var CAP_SORT={queen:5,rook:4,bishop:3,knight:2,pawn:1};
var CAP_SVG={pawn_w:'wP',pawn_b:'bP',knight_w:'wN',knight_b:'bN',bishop_w:'wB',bishop_b:'bB',rook_w:'wR',rook_b:'bR',queen_w:'wQ',queen_b:'bQ'};
var PIECE_VAL={pawn:1,knight:3,bishop:3,rook:5,queen:9};
var PROMO_W=['Q','R','N','B'], PROMO_B=['q','r','n','b'];
var PROMO_MAP={'Q':'q','R':'r','N':'n','B':'b','q':'q','r':'r','n':'n','b':'b'};

var flipped=false, selSq=null, legalTgts=[], lastFrom=null, lastTo=null;
var fen='', gameOver=false, playerColor='white', moveHist=[], curEval=0;
var dragSq=null, isThinking=false, animating=false;
var SZ=70;

document.addEventListener('DOMContentLoaded',function(){buildBoard();fetchState();});

function buildBoard(){
  var b=document.getElementById('board');
  b.innerHTML='';
  for(var row=0;row<8;row++){
    for(var col=0;col<8;col++){
      var r=flipped?row:7-row, c=flipped?7-col:col;
      var sq=document.createElement('div');
      var light=(r+c)%2===1;
      sq.className='sq '+(light?'light':'dark');
      sq.dataset.sq=String.fromCharCode(97+c)+(r+1);
      (function(s){
        sq.addEventListener('click',function(){clickSq(s);});
        sq.addEventListener('dragover',function(e){e.preventDefault();sq.classList.add('drag-over');});
        sq.addEventListener('dragleave',function(){sq.classList.remove('drag-over');});
        sq.addEventListener('drop',function(e){e.preventDefault();sq.classList.remove('drag-over');dropPiece(s);});
      })(sq.dataset.sq);
      b.appendChild(sq);
    }
  }
  buildCoords();
}

function buildCoords(){
  var rh='',fh='';
  for(var i=0;i<8;i++){
    var r=flipped?i+1:8-i;
    rh+='<div class="coord-r" style="top:'+(i*SZ)+'px;height:'+SZ+'px">'+r+'</div>';
  }
  var files='abcdefgh';
  for(var i=0;i<8;i++){
    var c=flipped?7-i:i;
    fh+='<div class="coord-f" style="left:'+(16+i*SZ)+'px;width:'+SZ+'px">'+files[c]+'</div>';
  }
  document.getElementById('coordR').innerHTML=rh;
  document.getElementById('coordF').innerHTML=fh;
}

function parseFEN(f){
  var rows=f.split(' ')[0].split('/');
  var pcs={};
  for(var r=0;r<8;r++){
    var col=0;
    for(var j=0;j<rows[r].length;j++){
      var ch=rows[r][j];
      if(ch>='1'&&ch<='8') col+=parseInt(ch);
      else{pcs[String.fromCharCode(97+col)+(8-r)]=ch;col++;}
    }
  }
  return pcs;
}

function render(inCheck,checkSide,hideSquares){
  // hideSquares: optional array of square names to skip placing pieces on
  var hide=hideSquares||[];
  var pcs=parseFEN(fen);
  var sqs=document.querySelectorAll('.sq');
  for(var i=0;i<sqs.length;i++){
    var sq=sqs[i];
    var s=sq.dataset.sq;
    sq.classList.remove('selected','last-move','in-check','drag-over');
    var old=sq.querySelectorAll('.pc,.dot,.ring');
    for(var j=0;j<old.length;j++) old[j].remove();

    if(s===lastFrom||s===lastTo) sq.classList.add('last-move');
    if(s===selSq) sq.classList.add('selected');

    if(hide.indexOf(s)>=0){/* skip piece on this square */}
    else{
      var pc=pcs[s];
      if(pc){
        var el=document.createElement('div');
        el.className='pc';
        el.innerHTML=SVG[P_SVG[pc]];
        el.draggable=true;
        (function(ss){
          el.addEventListener('dragstart',function(e){startDrag(e,ss);});
          el.addEventListener('dragend',endDrag);
        })(s);
        sq.appendChild(el);
        if(inCheck&&(pc==='K'||pc==='k')){
          var c2=pc==='K'?'white':'black';
          if(c2===checkSide) sq.classList.add('in-check');
        }
      }
    }

    if(selSq&&legalTgts.indexOf(s)>=0){
      if(pcs[s]){
        var ring=document.createElement('div');ring.className='ring';sq.appendChild(ring);
      }else{
        var dot=document.createElement('div');dot.className='dot';sq.appendChild(dot);
      }
    }
  }
}

function sqToPixel(sqStr){
  var file=sqStr.charCodeAt(0)-97;
  var rank=parseInt(sqStr[1])-1;
  var col=flipped?7-file:file;
  var row=flipped?rank:7-rank;
  var boardEl=document.getElementById('board');
  var rect=boardEl.getBoundingClientRect();
  return {
    x: rect.left + col*SZ + (SZ-56)/2,
    y: rect.top + row*SZ + (SZ-56)/2
  };
}

function animateMove(fromSq,toSq,pieceSvgKey,duration){
  // Returns {promise, remove} — promise resolves when animation ends,
  // but the overlay stays visible until remove() is called.
  // This prevents the flash between overlay removal and render().
  var el=document.createElement('div');
  var from=sqToPixel(fromSq);
  var to=sqToPixel(toSq);
  el.className='pc-anim';
  el.innerHTML=SVG[pieceSvgKey];
  el.style.position='fixed';
  el.style.left=from.x+'px';
  el.style.top=from.y+'px';
  el.style.transition='none';
  document.body.appendChild(el);

  var p=new Promise(function(resolve){
    requestAnimationFrame(function(){
      requestAnimationFrame(function(){
        el.style.transition='left '+duration+'ms cubic-bezier(.25,.1,.25,1), top '+duration+'ms cubic-bezier(.25,.1,.25,1)';
        el.style.left=to.x+'px';
        el.style.top=to.y+'px';
        var done=false;
        function finish(){if(!done){done=true;resolve();}}
        el.addEventListener('transitionend',finish);
        // Fallback in case transitionend doesn't fire
        setTimeout(finish,duration+50);
      });
    });
  });
  return {promise:p, remove:function(){el.remove();}};
}

function startDrag(e,sq){
  if(gameOver||isThinking||animating) return;
  dragSq=sq;
  setTimeout(function(){if(e.target.classList) e.target.classList.add('ghost');},0);
  e.dataTransfer.effectAllowed='move';
  getLegal(sq);
}
function endDrag(e){if(e.target.classList) e.target.classList.remove('ghost');dragSq=null;}
function dropPiece(to){
  if(dragSq&&dragSq!==to) tryMove(dragSq,to);
  clearSel();
}

function clickSq(sq){
  if(gameOver||isThinking||animating) return;
  if(selSq){
    if(selSq===sq){clearSel();render(false,null);return;}
    if(legalTgts.indexOf(sq)>=0){tryMove(selSq,sq);return;}
  }
  selSq=sq;getLegal(sq);
}
function clearSel(){selSq=null;legalTgts=[];}

function getLegal(sq){
  fetch('/api/legalmoves?square='+sq).then(function(r){return r.json();}).then(function(d){
    legalTgts=d.moves||[];
    if(!legalTgts.length) selSq=null; else selSq=sq;
    render(false,null);
  }).catch(function(){});
}

function tryMove(from,to){
  var pcs=parseFEN(fen);
  var pc=pcs[from];
  if(pc&&(pc==='P'||pc==='p')){
    var tr=parseInt(to[1]);
    if((pc==='P'&&tr===8)||(pc==='p'&&tr===1)){showPromo(from,to,pc==='P'?'white':'black');return;}
  }
  doMove(from,to,'');
}

function showPromo(from,to,color){
  var dlg=document.getElementById('promo');
  var pieces=color==='white'?PROMO_W:PROMO_B;
  dlg.innerHTML='';
  pieces.forEach(function(pc){
    var el=document.createElement('div');
    el.className='promo-item';
    el.innerHTML=SVG[P_SVG[pc]];
    el.onclick=function(){dlg.classList.remove('on');doMove(from,to,PROMO_MAP[pc]);};
    dlg.appendChild(el);
  });
  var board=document.getElementById('board');
  var sqs=board.querySelectorAll('.sq');
  var tgt=null;
  for(var i=0;i<sqs.length;i++){if(sqs[i].dataset.sq===to) tgt=sqs[i];}
  if(tgt){
    var br=board.getBoundingClientRect();
    var tr=tgt.getBoundingClientRect();
    dlg.style.left=(tr.left-br.left)+'px';
    dlg.style.top=(tr.top-br.top)+'px';
  }
  dlg.classList.add('on');
}

function doMove(from,to,promo){
  clearSel();
  animating=true;

  // Determine piece being moved for animation
  var pcs=parseFEN(fen);
  var pc=pcs[from];
  var svgKey=pc?P_SVG[pc]:null;

  // Phase 1: Animate the player's piece sliding from -> to
  // Hide the piece at origin AND any piece at destination (capture target)
  var playerAnim=null;
  if(svgKey){
    render(false,null,[from,to]);
    playerAnim=animateMove(from,to,svgKey,140);
  }

  var afterPlayerAnim=playerAnim?playerAnim.promise:Promise.resolve();

  afterPlayerAnim.then(function(){
    // Animation done. Build a local FEN with the piece moved so render
    // shows the piece at destination immediately, then remove overlay.
    // This eliminates the flash: real piece appears BEFORE overlay goes away.
    var localPcs=parseFEN(fen);
    var movedPc=localPcs[from];
    if(movedPc){
      delete localPcs[from];
      localPcs[to]=movedPc;
      // Build a temporary FEN string for rendering
      fen=buildFENFromPcs(localPcs,fen);
    }
    lastFrom=from; lastTo=to;
    render(false,null);
    // NOW remove the overlay — the real piece is already drawn underneath
    if(playerAnim) playerAnim.remove();

    setThinking(true);
    return fetch('/api/move',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({from:from,to:to,promo:promo})});
  }).then(function(r){return r.json();}).then(function(d){
    setThinking(false);
    if(!d.ok){animating=false;fetchState();return;}

    // Phase 2: Show the board after player's move (server-confirmed position)
    if(d.playerFen){
      fen=d.playerFen;
      lastFrom=d.playerFrom||from;
      lastTo=d.playerTo||to;
      updateCaptures(d.playerCaptured);
      render(false,null);
    }

    // If player's move ended the game
    if(d.playerGameOver){
      fen=d.fen;gameOver=true;
      if(d.moveList) moveHist=d.moveList;
      updateMoves();updateCaptures(d.captured);
      render(false,null);
      updateStatus(d);animating=false;
      return;
    }

    // Phase 3: Animate engine's reply
    if(d.engineFrom&&d.engineTo){
      var engPcs=parseFEN(fen);
      var engPc=engPcs[d.engineFrom];
      var engSvg=engPc?P_SVG[engPc]:null;

      setTimeout(function(){
        if(engSvg){
          render(false,null,[d.engineFrom,d.engineTo]);
          var engAnim=animateMove(d.engineFrom,d.engineTo,engSvg,180);
          engAnim.promise.then(function(){
            // Update to final state, render, THEN remove overlay
            lastFrom=d.engineFrom;lastTo=d.engineTo;
            fen=d.fen;gameOver=d.gameOver;
            if(d.moveList) moveHist=d.moveList;
            updateMoves();
            if(d.eval!==undefined) updateEval(d.eval);
            updateStats(d);updateCaptures(d.captured);
            var side=d.fen.split(' ')[1]==='w'?'white':'black';
            render(false,side);
            engAnim.remove();
            updateStatus(d);
            animating=false;
          });
        }else{
          lastFrom=d.engineFrom;lastTo=d.engineTo;
          fen=d.fen;gameOver=d.gameOver;
          if(d.moveList) moveHist=d.moveList;
          updateMoves();
          if(d.eval!==undefined) updateEval(d.eval);
          updateStats(d);updateCaptures(d.captured);
          render(false,null);
          updateStatus(d);
          animating=false;
        }
      },80);
    }else{
      fen=d.fen;gameOver=d.gameOver;
      if(d.moveList) moveHist=d.moveList;
      updateMoves();
      if(d.eval!==undefined) updateEval(d.eval);
      updateStats(d);updateCaptures(d.captured);
      render(false,null);
      updateStatus(d);
      animating=false;
    }
  }).catch(function(){setThinking(false);animating=false;});
}

// Build a FEN position string from a pieces map, preserving the rest of the FEN
function buildFENFromPcs(pcs,oldFen){
  var parts=oldFen.split(' ');
  var rows=[];
  for(var rank=7;rank>=0;rank--){
    var empty=0;var row='';
    for(var file=0;file<8;file++){
      var sq=String.fromCharCode(97+file)+(rank+1);
      if(pcs[sq]){
        if(empty>0){row+=empty;empty=0;}
        row+=pcs[sq];
      }else{empty++;}
    }
    if(empty>0) row+=empty;
    rows.push(row);
  }
  parts[0]=rows.join('/');
  return parts.join(' ');
}

function setThinking(v){
  isThinking=v;
  document.getElementById('think').classList.toggle('on',v);
}

function fetchState(){
  fetch('/api/state').then(function(r){return r.json();}).then(function(d){
    fen=d.fen;gameOver=d.gameOver;
    if(d.moveList) moveHist=d.moveList;
    updateMoves();updateCaptures(d.captured);
    render(d.inCheck,d.side);
    updateStatusState(d);
    updatePlayerBars();
  }).catch(function(){});
}

function updateStatus(d){
  var m=document.getElementById('sMain'),s=document.getElementById('sSub');
  s.classList.remove('over');
  if(d.gameOver){
    s.classList.add('over');
    if(d.result==='white'){m.textContent='White wins!';s.textContent='Checkmate';}
    else if(d.result==='black'){m.textContent='Black wins!';s.textContent='Checkmate';}
    else{m.textContent='Draw';s.textContent='Game drawn';}
  }else{
    m.textContent=(d.fen.split(' ')[1]==='w'?'White':'Black')+' to move';s.textContent='';
  }
}
function updateStatusState(d){
  var m=document.getElementById('sMain'),s=document.getElementById('sSub');
  s.classList.remove('over');
  if(d.gameOver){
    s.classList.add('over');
    if(d.result==='white'){m.textContent='White wins!';s.textContent='Checkmate';}
    else if(d.result==='black'){m.textContent='Black wins!';s.textContent='Checkmate';}
    else{m.textContent='Draw';s.textContent='Game drawn';}
  }else{
    m.textContent=(d.side==='white'?'White':'Black')+' to move';s.textContent='';
  }
}

function updateEval(cp){
  curEval=cp;
  var ev=cp/100;
  var pct=Math.min(97,Math.max(3, 50+(2/(1+Math.exp(-ev*0.6))-1)*50));
  document.getElementById('evalFill').style.height=pct+'%';
  var label=(ev>=0?'+':'')+ev.toFixed(1);
  document.getElementById('evalBot').textContent=label;
  document.getElementById('evalTop').textContent='';
}

function updateStats(d){
  if(d.eval!==undefined){var ev=d.eval/100;document.getElementById('iEval').textContent=(ev>=0?'+':'')+ev.toFixed(2);}
  if(d.depth) document.getElementById('iDep').textContent=d.depth;
  if(d.nodes) document.getElementById('iNod').textContent=fmt(d.nodes);
  if(d.nps) document.getElementById('iNps').textContent=fmt(d.nps);
  else if(d.nodes&&d.timeMs>0) document.getElementById('iNps').textContent=fmt(Math.round(d.nodes*1000/d.timeMs));
}
function fmt(n){
  if(n>=1e6) return(n/1e6).toFixed(1)+'M';
  if(n>=1e3) return(n/1e3).toFixed(1)+'K';
  return ''+n;
}

function updateMoves(){
  var c=document.getElementById('mList');c.innerHTML='';
  for(var i=0;i<moveHist.length;i+=2){
    var num=Math.floor(i/2)+1;
    var row=document.createElement('div');row.className='mv';
    var n=document.createElement('span');n.className='mv-n';n.textContent=num+'.';row.appendChild(n);
    var w=document.createElement('span');w.className='mv-w';
    if(i===moveHist.length-1||i===moveHist.length-2) w.classList.add('latest');
    w.textContent=moveHist[i].uci;row.appendChild(w);
    if(i+1<moveHist.length){
      var b=document.createElement('span');b.className='mv-b';
      if(i+1===moveHist.length-1) b.classList.add('latest');
      b.textContent=moveHist[i+1].uci;row.appendChild(b);
    }
    c.appendChild(row);
  }
  c.scrollTop=c.scrollHeight;
}

function updateCaptures(cap){
  if(!cap) return;
  var wCap=(cap.white||[]).sort(function(a,b){return(CAP_SORT[b]||0)-(CAP_SORT[a]||0);});
  var bCap=(cap.black||[]).sort(function(a,b){return(CAP_SORT[b]||0)-(CAP_SORT[a]||0);});
  var wAdv=0,bAdv=0;
  wCap.forEach(function(p){wAdv+=(PIECE_VAL[p]||0);});
  bCap.forEach(function(p){bAdv+=(PIECE_VAL[p]||0);});
  var diff=wAdv-bAdv;

  function renderCaps(pieces,color){
    return pieces.map(function(p){
      var key=p+'_'+color;
      var svgKey=CAP_SVG[key];
      if(!svgKey) return '';
      return '<span style="display:inline-block;width:14px;height:14px;">'+SVG[svgKey].replace('viewBox="0 0 45 45"','viewBox="0 0 45 45" width="14" height="14"')+'</span>';
    }).join('');
  }

  var topEl=document.getElementById('topCap'), botEl=document.getElementById('botCap');
  var topAdv=document.getElementById('topAdv'), botAdv=document.getElementById('botAdv');

  if(playerColor==='white'){
    topEl.innerHTML=renderCaps(bCap,'w');
    botEl.innerHTML=renderCaps(wCap,'b');
    topAdv.textContent=diff<0?'+'+Math.abs(diff):'';
    botAdv.textContent=diff>0?'+'+diff:'';
  }else{
    topEl.innerHTML=renderCaps(wCap,'b');
    botEl.innerHTML=renderCaps(bCap,'w');
    topAdv.textContent=diff>0?'+'+diff:'';
    botAdv.textContent=diff<0?'+'+Math.abs(diff):'';
  }
}

function updatePlayerBars(){
  var topColor=document.getElementById('topColor');
  var botColor=document.getElementById('botColor');
  var topName=document.getElementById('topName');
  var botName=document.getElementById('botName');

  if(flipped){
    if(playerColor==='white'){
      topColor.className='p-color white-c';
      botColor.className='p-color black-c';
      topName.textContent='You (White)';
      botName.textContent='Engine';
    }else{
      topColor.className='p-color black-c';
      botColor.className='p-color white-c';
      topName.textContent='You (Black)';
      botName.textContent='Engine';
    }
  }else{
    if(playerColor==='white'){
      topColor.className='p-color black-c';
      botColor.className='p-color white-c';
      topName.textContent='Engine';
      botName.textContent='You (White)';
    }else{
      topColor.className='p-color white-c';
      botColor.className='p-color black-c';
      topName.textContent='Engine';
      botName.textContent='You (Black)';
    }
  }
}

function flip(){
  flipped=!flipped;
  buildBoard();render(false,null);
  updatePlayerBars();
  fetchState();
}

function showModal(){document.getElementById('modalBg').classList.add('on');}
function hideModal(){document.getElementById('modalBg').classList.remove('on');}

function newGame(){
  var color=document.getElementById('mColor').value;
  var depth=parseInt(document.getElementById('mDepth').value);
  hideModal();
  playerColor=color;
  animating=false;
  setThinking(true);
  fetch('/api/newgame',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({color:color,depth:depth})}).then(function(r){return r.json();}).then(function(d){
    setThinking(false);
    fen=d.fen;gameOver=false;
    lastFrom=d.engineFrom||null;lastTo=d.engineTo||null;
    moveHist=d.moveList||[];
    clearSel();updateMoves();updateCaptures(d.captured);
    if(d.eval!==undefined) updateEval(d.eval);
    updateStats(d);
    render(false,null);
    updatePlayerBars();
    document.getElementById('sMain').textContent=(d.fen.split(' ')[1]==='w'?'White':'Black')+' to move';
    document.getElementById('sSub').textContent='';
    document.getElementById('sSub').classList.remove('over');
    if(!d.eval){
      document.getElementById('evalFill').style.height='50%';
      document.getElementById('evalBot').textContent='0.0';
      document.getElementById('iEval').textContent='0.00';
      document.getElementById('iDep').textContent='\u2014';
      document.getElementById('iNod').textContent='\u2014';
      document.getElementById('iNps').textContent='\u2014';
    }
    toast('Game started \u2014 '+color+' \u2022 Depth '+depth);
  }).catch(function(){setThinking(false);});
}

function undo(){
  if(gameOver||isThinking||animating) return;
  fetch('/api/undo',{method:'POST'}).then(function(r){return r.json();}).then(function(d){
    fen=d.fen;gameOver=false;lastFrom=null;lastTo=null;
    moveHist=d.moveList||[];
    clearSel();updateMoves();updateCaptures(d.captured);
    render(false,null);
    document.getElementById('sMain').textContent=(d.fen.split(' ')[1]==='w'?'White':'Black')+' to move';
    document.getElementById('sSub').textContent='';
    document.getElementById('sSub').classList.remove('over');
  }).catch(function(){});
}

function toast(msg){
  var t=document.getElementById('toast');
  t.textContent=msg;t.classList.add('show');
  setTimeout(function(){t.classList.remove('show');},2200);
}

document.addEventListener('click',function(e){if(e.target.id==='modalBg') hideModal();});
document.addEventListener('keydown',function(e){
  if(e.key==='Escape'){hideModal();document.getElementById('promo').classList.remove('on');}
});
</script>
</body>
</html>` + ""
