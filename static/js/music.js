function displayDuration(duration) {
    const hours = Math.floor(duration / 3600);
    const minutes = Math.floor((duration % 3600) / 60);
    const seconds = Math.floor(duration % 60);

    let formattedDuration;
    if (hours > 0) {
        formattedDuration = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    } else {
        formattedDuration = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    }
    return formattedDuration
}

let players = []

class MusicPlayer {
    constructor(pid, url, duration) {
        let playerbox = $(`#${pid}`)
        let ringbtn = playerbox.find(".ring")
        this.ringbtn = ringbtn
        let progresstotal = playerbox.find(".progresstotal")
        let progressbar = playerbox.find(".progressbar")
        this.progressbar = progressbar
        this.cover = playerbox.find(".cover")
        let musictime = playerbox.find(".musictime")
        this.musictime = musictime

        this.playerbtn = $(`<div class="btn">
            <i class="iconfont icon-play-filling"></i>
            </div>`)
        this.pausebtn = $(`<div  class="btn">
            <i class="iconfont icon-pausecircle-fill"></i>
        </div>`)

        this.url = url

        let musicIsReady = (duration) => {
            musictime.text(displayDuration(duration))
            this.cover.append(this.playerbtn)
            this.playerbtn.click(()=>{
                this.play()
            })
            this.pausebtn.click(()=>{
                this.pause()
            })
        } 
        if(duration > 0){
            musicIsReady(duration)
        }else{
            this.initAudio()
            this.audio.addEventListener('loadedmetadata',  () => {
                musicIsReady(audio.duration)
            });
        }

        this.ringbtnIsPressed = false
       
        ringbtn.on("mousedown", ()=>{
            this.ringbtnIsPressed = true
        })
        playerbox.on("mousemove", event =>{
            let audio = this.audio
            if(this.ringbtnIsPressed){
                const progressRect = progresstotal[0].getBoundingClientRect();
                const offsetX = event.clientX - progressRect.left;
                const percentage = (offsetX / progressRect.width) * 100;
                if(percentage < 0){
                    percentage = 0
                }
                if(percentage > 100){
                    percentage = 100
                }
                ringbtn.css("left", percentage + "%")
                progressbar.css("width", percentage + "%")
                if(audio.duration > 0){
                    let settime = Math.floor(audio.duration * percentage / 100)
                    musictime.text(displayDuration(audio.duration - settime))
                    audio.setAttribute("setTime", settime)
                }
            }
        })

        playerbox.on("mouseleave mouseup", ()=>{
            let audio = this.audio
            if(this.ringbtnIsPressed){
                let setTime = audio.getAttribute("settime")
                if(setTime){
                    let setTime0 = parseInt(setTime)
                    audio.currentTime = setTime0;
                }

                this.ringbtnIsPressed = false
            }
            
        })

        players.push(this)

    }

    initAudio(){
        const audio = new Audio(this.url);
        this.audio = audio 
        
        audio.addEventListener('timeupdate', () => {
            if(!this.ringbtnIsPressed){
                this.musictime.text(displayDuration(audio.duration - audio.currentTime))
                let progress = (audio.currentTime / audio.duration * 100) + "%"
                this.ringbtn.css("left", progress)
                this.progressbar.css("width", progress)
            }
           
        });

        audio.addEventListener('ended',  () => {
            this.pause()
        });

        audio.addEventListener('error',  () => {
            toast('无法加载音频文件，请检查URL是否正确');
        });
    }

    pause(){
        if(this.audio){
            this.pausebtn.remove()
            this.playerbtn.click(()=>{
                this.play()
            })
            this.cover.append(this.playerbtn)
            this.audio.pause()
        }
        
    }

    play(){
        for(let i = 0; i < players.length; i++){
            players[i].pause()
        }

        if(!this.audio){
            this.initAudio()
        }

        this.playerbtn.remove()
        this.pausebtn.click(()=>{
            this.pause()
        })
        this.cover.append(this.pausebtn)
        this.audio.play()
    }
}