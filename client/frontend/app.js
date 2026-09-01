(() => {
"use strict";

const $ = id => document.getElementById(id);

const el = {
  welcome: $("welcome"), welcomeForm: $("welcomeForm"),
  name: $("name"), createMode: $("createMode"), joinMode: $("joinMode"),
  joinBox: $("joinBox"), joinCode: $("joinCode"), submit: $("submit"),
  advanced: $("advanced"), advancedBox: $("advancedBox"), server: $("server"),
  connection: $("connection"), roomName: $("roomName"),
  roomCode: $("roomCodeLarge"), roomCodeMini: $("roomCode"),
  copy: $("copy"), users: $("users"), role: $("role"),
  songForm: $("songForm"), url: $("url"), list: $("list"), count: $("count"),
  title: $("title"), subtitle: $("subtitle"), progress: $("progress"),
  current: $("current"), length: $("lengthBottom"), play: $("play"),
  back: $("back"), forward: $("forward"), stop: $("stop"), toast: $("toast"),
  trackState: $("trackState"), cover: $("cover"), trackNumber: $("trackNumber"),
  volume: $("volume"), volumeValue: $("volumeValue"), mute: $("mute"),
  globalVolume: $("globalVolume"), globalVolumeValue: $("globalVolumeValue"),
  globalMute: $("globalMute"), adminPanel: $("adminPanel"),
  membersList: $("membersList"), djPanel: $("djPanel"),
  historyList: $("historyList"), historyRefresh: $("historyRefresh")
};

const player = new Audio();
player.autoplay = false;
player.preload = "auto";
player.volume = 1.0;
player.muted = false;

player.addEventListener("loadedmetadata", () => {
  console.log("NT-MUSIC MEDIA LOADED:", {
    duration: player.duration,
    currentTime: player.currentTime,
    readyState: player.readyState,
    networkState: player.networkState,
    src: player.currentSrc
  });
});

player.addEventListener("canplay", () => {
  console.log("NT-MUSIC CAN PLAY:", {
    readyState: player.readyState,
    currentTime: player.currentTime,
    duration: player.duration
  });
});

player.addEventListener("playing", () => {
  console.log("NT-MUSIC PLAYING:", {
    currentTime: player.currentTime,
    duration: player.duration,
    volume: player.volume,
    muted: player.muted
  });
});

player.addEventListener("timeupdate", () => {
  console.log("NT-MUSIC TIME:", player.currentTime);
});

player.addEventListener("waiting", () => {
  console.log("NT-MUSIC WAITING");
});

player.addEventListener("stalled", () => {
  console.log("NT-MUSIC STALLED");
});

player.addEventListener("error", () => {
  console.error("NT-MUSIC MEDIA ERROR:", player.error);
});

player.addEventListener("error", () => {
  const err = player.error;
  const message = err ? "MEDIA ERROR CODE " + err.code : "UNKNOWN MEDIA ERROR";
  console.error("NT-MUSIC MEDIA ERROR:", err, "URL:", player.src);
  toast(message);
});

player.addEventListener("playing", () => {
  console.log("NT-MUSIC AUDIO IS PLAYING");
  toast("Audio started.");
});

player.addEventListener("canplay", () => {
  console.log("NT-MUSIC AUDIO CAN PLAY");
});

const state = {
  ws:null,
  mode:"create",
  userId:(crypto.randomUUID ? crypto.randomUUID() : Date.now()+"-"+Math.random()),
  name:localStorage.getItem("ntmusic-name") || "",
  server:`${location.protocol==="https:"?"wss:":"ws:"}//${location.host}/ws`,
  room:"",
  owner:"",
  users:0,
  queue:[],
  history:[],
  members:[],
  role:"member",
  volume:1,
  globalVolume:1,
  globalMuted:false,
  localMuted:false,
  playback:{song_id:"",playing:false,position:0,updated:0},
  received:Date.now(),
  heartbeat:null
};

el.name.value=state.name;
el.server.value=state.server;

function toast(text){
  el.toast.textContent=text;
  el.toast.classList.add("show");
  clearTimeout(toast.t);
  toast.t=setTimeout(()=>el.toast.classList.remove("show"),2800);
}

function mode(join){
  state.mode=join?"join":"create";
  el.createMode.classList.toggle("active",!join);
  el.joinMode.classList.toggle("active",join);
  el.joinBox.classList.toggle("hide",!join);
  el.submit.textContent=join?"Join room →":"Create room →";
}
el.createMode.addEventListener("click",()=>mode(false));
el.joinMode.addEventListener("click",()=>mode(true));

function connect(){
  if(state.ws) try{state.ws.close()}catch{}
  state.server=(el.server.value||state.server).trim();
  if(state.server.startsWith("http://"))state.server="ws://"+state.server.slice(7);
  if(state.server.startsWith("https://"))state.server="wss://"+state.server.slice(8);
  if(!state.server.startsWith("ws://")&&!state.server.startsWith("wss://"))state.server="ws://"+state.server;
  if(!state.server.endsWith("/ws"))state.server=state.server.replace(/\/+$/,"")+"/ws";
  localStorage.setItem("ntmusic-server",state.server);

  try{state.ws=new WebSocket(state.server)}catch{toast("Could not connect to the server.");return}

  state.ws.addEventListener("open",()=>{
    el.connection.className="online";
    el.connection.innerHTML="<i></i> Connected";
    const data={user_id:state.userId,name:state.name};
    if(state.mode==="join")data.room_code=el.joinCode.value.trim().toUpperCase();
    send(state.mode==="join"?"join_room":"create_room",data);
  });

  state.ws.addEventListener("message",e=>{
    let m;
    try{m=JSON.parse(e.data)}catch{return}
    handle(m);
  });

  state.ws.addEventListener("close",()=>{
    el.connection.className="offline";
    el.connection.innerHTML="<i></i> Offline";
  });

  state.ws.addEventListener("error",()=>toast("WebSocket connection error."));
  clearInterval(state.heartbeat);
  state.heartbeat=setInterval(()=>{
    if(state.ws&&state.ws.readyState===WebSocket.OPEN)send("heartbeat");
  },300000);
}

function send(type,data={}){
  if(!state.ws||state.ws.readyState!==WebSocket.OPEN){
    toast("Not connected to the room.");
    return false;
  }
  state.ws.send(JSON.stringify({type,data}));
  return true;
}

function dataOf(m){return m&&m.data!==undefined?m.data:m}

function handle(m){
  const d=dataOf(m)||{};
  switch(m.type){
    case "room_created":
      state.room=d.code||d.room_code||d.room||d.Code||"";
      state.owner=state.userId;
      state.users=Number(d.user_count||d.users||1);
      state.role="owner";
      enter();
      render();
      send("sync_state");
      toast("Room created.");
      break;

    case "room_joined":
      state.room=d.code||d.room_code||d.room||d.Code||"";
      state.owner=d.owner_id||d.owner||d.OwnerID||"";
      state.users=Number(d.user_count||d.users||1);
      enter();
      toast("Joined room.");
      break;

    case "queue_state":
      state.queue=Array.isArray(d)?d:(d.queue||d.songs||d.items||[]);
      render();
      break;

    case "song_added":
      {
        const song=d.song||d;
        const id=songId(song);
        if(id&&!state.queue.some(x=>songId(x)===id))state.queue.push(song);
        else if(!id)state.queue.push(song);
        render();
      }
      break;

    case "playback_state":
      state.playback = {
        song_id: String(
          d.song_id ??
          d.songId ??
          d.SongID ??
          d.id ??
          ""
        ),
        playing: Boolean(
          d.playing ??
          d.Playing ??
          false
        ),
        position: Math.max(
          0,
          Number(d.position ?? d.Position ?? 0)
        ),
        updated: Number(
          d.updated ??
          d.Updated ??
          Date.now()
        )
      };

      state.received = Date.now();

      {
        const current = state.queue.find(
          song => songId(song) === state.playback.song_id
        );

        if (current && songUrl(current)) {
          const audioSource =
            `${location.origin}/audio?url=` +
            encodeURIComponent(songUrl(current));

          const targetPosition = Math.max(
            0,
            Number(state.playback.position) || 0
          );

          const currentAudioSongId = player.dataset.ntMusicSongId || "";

          if (currentAudioSongId !== state.playback.song_id) {
            player.pause();
            player.dataset.ntMusicSongId = state.playback.song_id;
            player.src = audioSource;
            player.load();

            player.addEventListener("loadedmetadata", () => {
              if (player.dataset.ntMusicSongId !== state.playback.song_id) return;

              try {
                player.currentTime = targetPosition;
              } catch {}

              if (state.playback.playing) {
                player.play().catch(error => {
                  console.error("NT-MUSIC AUDIO PLAY ERROR:", error);
                });
              }
            }, { once: true });
          } else {
            const drift =
              Math.abs(
                (Number(player.currentTime) || 0) -
                targetPosition
              );

            if (drift > 1.5) {
              try {
                player.currentTime = targetPosition;
              } catch {}
            }

            if (state.playback.playing) {
              player.play().catch(() => {});
            } else {
              player.pause();
            }
          }
        } else if (!state.playback.song_id) {
          player.pause();
        }
      }

      render();
      break;

    case "history_state":
      state.history=Array.isArray(d)?d:(d.songs||d.items||[]);
      renderHistory();
      break;

    case "permissions_state":
      state.members=Array.isArray(d)?d:(d.users||d.members||[]);

      const me=state.members.find(
        x=>String(x.user_id||"")===String(state.userId)
      );

      if(state.owner===state.userId){
        state.role="owner";
      }else if(me){
        state.role=String(me.role||"member");
      }

      state.users=state.members.length||state.users;
      renderMembers();
      render();
      break;

    case "audio_state":
      state.volume=Math.max(0,Math.min(1,Number(d.user_volume??1)));
      state.globalVolume=Math.max(0,Math.min(1,Number(d.global_volume??1)));
      state.globalMuted=Boolean(d.global_muted);
      player.volume=Math.max(
        0,
        Math.min(
          1,
          Number(d.effective_volume ??
            state.volume * state.globalVolume)
        )
      );
      player.muted=state.globalMuted||state.localMuted;
      render();
      break;

    case "error":
      toast(d.message||d.error||d.Message||"Server rejected the request.");
      break;
  }
}

function enter(){
  el.welcome.classList.add("hide");
  render();
}

function songId(s){
  return String(s?.id??s?.ID??s?.song_id??s?.songId??"");
}
function songTitle(s){
  return String(s?.title??s?.Title??s?.name??s?.Name??s?.url??s?.URL??"Untitled song");
}
function songUrl(s){return String(s?.url??s?.URL??"")}

function position(){
  let p=Number(state.playback.position)||0;
  if(state.playback.playing&&state.playback.updated){
    p+=(Date.now()-Number(state.playback.updated))/1000;
  }
  return Math.max(0,p);
}
function time(v){
  v=Math.max(0,Math.floor(Number(v)||0));
  return Math.floor(v/60)+":"+String(v%60).padStart(2,"0");
}

function owner(){return !!state.owner&&state.owner===state.userId}
function canControl(){return owner()||state.role==="dj"}

function renderHistory(){
  if(!el.historyList)return;

  if(!state.history.length){
    el.historyList.innerHTML=`
      <div class="emptyState historyEmpty">
        <div class="emptySymbol historySymbol">♪</div>
        <div class="emptyTitle">NO LISTENING HISTORY</div>
        <div class="emptyText">Played tracks will appear here.</div>
      </div>
    `;
    return;
  }

  el.historyList.innerHTML=state.history.map(s=>`
    <div class="historyRow">
      <div class="historyIcon">♪</div>
      <div class="historyTitle">${esc(songTitle(s))}</div>
      <button class="historyPlay" data-song-id="${esc(songId(s))}">PLAY</button>
    </div>
  `).join("");

  el.historyList.querySelectorAll(".historyPlay").forEach(b=>{
    b.addEventListener("click",()=>{
      if(canControl())send("play_history",{song_id:b.dataset.songId});
      else toast("Only the owner or DJ can play history.");
    });
  });
}

function renderMembers(){
  if(!el.membersList)return;

  let members=Array.isArray(state.members)?state.members.slice():[];

  if(owner() && !members.some(
    member=>String(member.user_id||"")===String(state.userId)
  )){
    members.unshift({
      user_id:state.userId,
      name:state.name||"You",
      role:"owner"
    });
  }

  if(!members.length){
    el.membersList.innerHTML='<div class="empty compact"><span>No members connected.</span></div>';
    return;
  }

  el.membersList.innerHTML=members.map(member=>{
    const id=String(member.user_id||"");
    const role=String(member.role||"member");

    const button=owner()&&role!=="owner"
      ? `<button class="memberButton" data-user-id="${esc(id)}" data-role="${esc(role)}">${role==="dj"?"REMOVE DJ":"MAKE DJ"}</button>`
      : "";

    return `
      <div class="memberRow">
        <div>
          <div class="memberName">${esc(member.name||id)}</div>
          <div class="memberRole">${esc(role.toUpperCase())}</div>
        </div>
        ${button}
      </div>
    `;
  }).join("");

  el.membersList.querySelectorAll(".memberButton").forEach(button=>{
    button.addEventListener("click",()=>{
      const id=button.dataset.userId;
      if(!id)return;

      if(button.dataset.role==="dj"){
        send("remove_dj",{target_user_id:id});
      }else{
        send("set_dj",{target_user_id:id});
      }
    });
  });
}

function render(){
  if(el.adminPanel){
    const showAdmin = owner();
    el.adminPanel.classList.toggle("hide", !showAdmin);
    el.adminPanel.style.display = showAdmin ? "block" : "none";
  }

  if(el.djPanel){
    const showDJ = state.role === "dj" && !owner();
    el.djPanel.classList.toggle("hide", !showDJ);
    el.djPanel.style.display = showDJ ? "block" : "none";
  }

  el.roomName.textContent=state.room?"Listening room":"Not connected";
  el.roomCode.textContent=state.room||"----";
  if(el.roomCodeMini)el.roomCodeMini.textContent=state.room||"----";
  el.users.textContent=state.users;
  el.role.textContent=state.role==="owner"?"OWNER":state.role==="dj"?"DJ":state.room?"MEMBER":"WAITING";

  el.count.textContent=state.queue.length;

  if(!state.queue.length){
    el.list.innerHTML=`
      <div class="emptyState roomEmpty">
        <div class="emptySymbol">♪</div>
        <div class="emptyTitle">QUEUE IS EMPTY</div>
        <div class="emptyText">Add a YouTube song to begin listening.</div>
      </div>
    `;
  }else{
    el.list.innerHTML=state.queue.map((s,i)=>`
      <div class="queueRow" data-song-id="${esc(songId(s))}" tabindex="0">
        <div class="queueNum">${i+1}</div>
        <div>
          <div class="queueTitle">${esc(songTitle(s))}</div>
          <div class="queueUrl">${esc(songUrl(s))}</div>
        </div>
        <div class="queueNum">${songId(s)===String(state.playback.song_id)?"PLAYING":"PLAY"}</div>
      </div>`).join("");

    el.list.querySelectorAll(".queueRow").forEach(row=>{
      const selectSong=()=>{
        if(!canControl()){
          toast("Only the owner or DJ can change the song.");
          return;
        }
        send("play_song",{song_id:row.dataset.songId});
      };

      row.addEventListener("click",selectSong);
      row.addEventListener("keydown",e=>{
        if(e.key==="Enter"||e.key===" "){
          e.preventDefault();
          selectSong();
        }
      });
    });
  }

  const current=state.queue.find(s=>songId(s)===String(state.playback.song_id));
  el.title.textContent=current?songTitle(current):"Nothing playing";
  el.subtitle.textContent=current?"Queued in this room":"Add a song to the queue.";

  const p=position();
  el.progress.value=Math.min(p,600);
  el.current.textContent=time(p);

  const enabled=canControl()&&state.ws&&state.ws.readyState===WebSocket.OPEN;
  el.play.disabled=!enabled;
  el.back.disabled=!enabled;
  el.forward.disabled=!enabled;
  el.stop.disabled=!enabled;
  el.progress.disabled=!enabled;
  el.play.textContent=state.playback.playing?"Ⅱ":"▶";

  const img = current?.thumbnail;

  if(el.cover){
    if(img){
      el.cover.style.backgroundImage=`url("${String(img).replace(/"/g,'\\\"')}")`;
      el.cover.style.backgroundSize="cover";
      el.cover.style.backgroundPosition="center";
      el.cover.style.backgroundRepeat="no-repeat";

      const icon=el.cover.querySelector("span");
      if(icon)icon.style.display="none";
    }else{
      el.cover.style.backgroundImage="";
      const icon=el.cover.querySelector("span");
      if(icon)icon.style.display="";
    }
  }
}

function esc(v){
  return String(v).replaceAll("&","&amp;").replaceAll("<","&lt;").replaceAll(">","&gt;").replaceAll('"',"&quot;");
}

el.welcomeForm.addEventListener("submit",e=>{
  e.preventDefault();
  state.name=el.name.value.trim();
  if(!state.name){toast("Enter your name.");return}
  if(state.mode==="join"&&!el.joinCode.value.trim()){toast("Enter the room code.");return}
  localStorage.setItem("ntmusic-name",state.name);
  connect();
});

el.songForm.addEventListener("submit",e=>{
  e.preventDefault();
  const url=el.url.value.trim();
  if(!url){toast("Paste a music URL.");return}
  if(send("add_song",{url}))el.url.value="";
});

el.play.addEventListener("click",()=>{
  if(!canControl())return;

  if(state.playback.playing){
    player.pause();
    send("pause");
    return;
  }

  const current=state.queue.find(s=>songId(s)===String(state.playback.song_id));

  if(current&&songUrl(current)){
    const audioSource = `${location.origin}/audio?url=` + encodeURIComponent(songUrl(current));

    if(player.src !== audioSource){
      player.src = audioSource;
      player.currentTime = state.playback.position || 0;
      player.load();
    }

    player.play().catch(e=>{
      console.error("NT-MUSIC AUDIO PLAY ERROR:", e);
      toast("AUDIO ERROR: " + e.name + " - " + e.message);
    });
  }

  send("play");
});
el.stop.addEventListener("click",()=>canControl()&&send("stop"));
el.back.addEventListener("click",()=>canControl()&&send("seek",{position:Math.max(0,position()-10)}));
el.forward.addEventListener("click",()=>canControl()&&send("seek",{position:position()+10}));
el.progress.addEventListener("change",()=>canControl()&&send("seek",{position:Number(el.progress.value)}));

document.addEventListener("keydown",e=>{
  if(e.target.tagName==="INPUT")return;
  if(e.code==="Space"){e.preventDefault();if(owner())send(state.playback.playing?"pause":"play")}
  if(e.key==="ArrowLeft"){e.preventDefault();if(owner())send("seek",{position:Math.max(0,position()-5)})}
  if(e.key==="ArrowRight"){e.preventDefault();if(owner())send("seek",{position:position()+5})}
});

el.copy.addEventListener("click",async()=>{
  if(!state.room)return;
  try{await navigator.clipboard.writeText(state.room);toast("Room code copied.")}catch{toast("Could not copy room code.")}
});

el.advanced.addEventListener("click",()=>el.advancedBox.classList.toggle("hide"));

el.volume?.addEventListener("input",()=>{
  const value=Math.max(0,Math.min(1,Number(el.volume.value)/100));

  state.volume=value;
  player.volume=value*state.globalVolume;
  player.muted=state.globalMuted||state.localMuted;

  if(el.volumeValue){
    el.volumeValue.textContent=Math.round(value*100)+"%";
  }

  send("set_volume",{volume:value});
});

el.mute?.addEventListener("click",()=>{
  player.muted=!player.muted;
  el.mute.textContent=player.muted?"UNMUTE MY AUDIO":"MUTE MY AUDIO";
});

el.globalVolume?.addEventListener("input",()=>{
  if(!owner())return;

  const value=Math.max(
    0,
    Math.min(1,Number(el.globalVolume.value)/100)
  );

  state.globalVolume=value;
  player.volume=state.volume*value;

  if(el.globalVolumeValue){
    el.globalVolumeValue.textContent=Math.round(value*100)+"%";
  }

  send("set_global_volume",{global_volume:value});
});

el.globalMute?.addEventListener("click",()=>{
  if(!owner())return;

  send("set_global_mute",{
    enabled:!state.globalMuted
  });
});

el.historyRefresh?.addEventListener("click",()=>send("history"));

player.addEventListener("ended",()=>{
  if(owner())send("next_song");
});

setInterval(()=>{if(state.playback.playing)render()},500);
mode(false);
render();
})();
