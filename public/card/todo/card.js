let time = model.endTime
if (time) {
    let endtime = new Date(time).getTime()
    let now = new Date().getTime()
    let sub = parseInt((endtime - now) / 1000)
    let countdown = carddiv.find(".countdown")
    if (sub < 0) {
        countdown.html(`<span class="finished">事项已经截止</span>`)
    } else {
        let days = parseInt(sub / 86400)
        let hours = parseInt(sub % 86400 / 3600)
        let color = "#67C23A"
        if (sub < 86400) {
            color = "#F56C6C"
        } else if (sub < 86400 * 3) {
            color = "#E6A23C"
        }

        countdown.html(`<span class="bigscreen">距离事项结束还有</span> <span class="bigtext" style="color:${color}"> ${days} </span>天<span  class="bigtext" style="color:${color}"> ${hours} </span>小时`)
    }
    carddiv.find(".endTimeText").text(time.replace("T", " "))
}
