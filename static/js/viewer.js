class ImageViewer {

    constructor() {

        this.close()

        this.prebtn = $("#image-viewer").find(".image-viewer__prev")
        this.nextbtn = $("#image-viewer").find(".image-viewer__next")
        this.closebtn = $("#image-viewer").find(".image-viewer__close")

        this.mainvideo = $("#image-viewer").find("#video")
        this.mainimage = $("#image-viewer").find("#image")

        this.waiticon =   $("#image-viewer").find(".waiticon")

        this.prebtn.click(() => {
            this.pre()
        })

        this.nextbtn.click(() => {
            this.next()
        })

        this.closebtn.click(() => {
            this.close()
        })
    }

    initViewer(parendid, curindex, mode) {

        let playmode = "image"
        if (mode === "video") {
            playmode = "video"
        }

        this.playmode = playmode

        this.curindex = curindex;
        this.images = []

        let imgs = $(`#${parendid}`).find("img")
        if (playmode === "image") {
            // 图片模式
            this.mainimage.css("display", "block")
            this.mainvideo.css("display", "none")
            this.main = this.mainimage
        } else {
            // 视频模式
            this.mainimage.css("display", "none")
            this.mainvideo.css("display", "block")
            this.main = this.mainvideo
        }

        for (let i = 0; i < imgs.length; i++) {
            let thumbUrl = imgs[i].getAttribute("src")
            let imageurl = imgs[i].getAttribute("url")
            this.images.push({
                thumb: thumbUrl,
                url: imageurl
            })
        }

        this.setView(this.images[this.curindex])
        this.updateButtonStatus()
        this.show()
    }

    setView(image) {
        let thumb = image.thumb
        let url = image.url
        if (this.playmode === "image") {
            this.main.attr("src", "")
            this.main.attr("src", thumb)
            this.waiticon.css("display","block")

            let realimg = new Image()
            realimg.crossOrigin="anonymous"
            realimg.onload = () => {
                var canvas = document.createElement('canvas');
                var context = canvas.getContext('2d');
                canvas.height = realimg.naturalHeight;
                canvas.width = realimg.naturalWidth;
                context.drawImage(realimg, 0, 0);
                var dataURL = canvas.toDataURL('image/jpeg');
                this.main.attr("src", dataURL)
                this.waiticon.css("display","none")
            }
            realimg.src = url
        } else if (this.playmode === "video") {
            this.main.attr("src", url)
        }
    }

    updateButtonStatus() {
        this.prebtn.css("display", "flex")
        this.nextbtn.css("display", "flex")

        if (this.curindex === 0) {
            this.prebtn.css("display", "none")
        }
        if (this.curindex === this.images.length - 1) {
            this.nextbtn.css("display", "none")
        }
    }

    pre() {
        if (this.curindex > 0) {
            this.curindex = this.curindex - 1
            this.setView(this.images[this.curindex])
            this.updateButtonStatus();
        }
    }

    next() {
        if (this.curindex < this.images.length - 1) {
            this.curindex = this.curindex + 1
            this.setView(this.images[this.curindex])
            this.updateButtonStatus();
        }
    }

    show() {
        $("#image-viewer").css("display", "block")
    }

    close() {
        if (this.playmode === "video") {
            this.main[0].pause();
            this.main.attr("src", "")
        }
        $("#image-viewer").css("display", "none")
    }
}