var imageViewer;

window.addEventListener("DOMContentLoaded", async () => {

    checkSiteStatus(async () => {
        checkLoginStatus(async () => {
            imageViewer = new ImageViewer();
            if (loginuser) {
                initEditPage()
                initSettingPage()
            }
            initBanner()
            initLoadMoreBtn()
            initGoTopBtn();
        })

    })
})

let footerIsFixed = true
window.addEventListener('scroll', () => {
    const scrollY = window.scrollY || document.documentElement.scrollTop;
    const windowHeight = window.innerHeight;

    if (scrollY >= 1.3 * windowHeight && footerIsFixed) {
        $(".footer").css("display", "none")
        $(".innter-footer").css("display", "block")
        footerIsFixed = false
    }

    if (scrollY >= 1.5 * windowHeight) {
        $("#gotop").css("display", "block")
    } else {
        $("#gotop").removeClass("active")
        $("#gotop").css("display", "none")
    }
});

function initGoTopBtn() {
    $("#gotop").on('mouseover touchstart', function () {
        $("#gotop").addClass("active")
    })
    $("#gotop").on('mouseout touchend', function () {
        $("#gotop").removeClass("active")
    })
    $("#gotop").on("click", () => {
        scrollToTop()
    })
}

async function checkLoginStatus(callback) {
    callback = callback || function () { }
    let authCheckResp = await fetch("/api/v1/user/check", {
        method: "POST"
    })

    if (authCheckResp.status === 200) {
        let user = (await authCheckResp.json()).data
        loginuser = user
    }

    callback()
}

var loginuser;

async function checkSiteStatus(successcallback) {
    successcallback = successcallback || function () { }
    let checkResp = await fetch("/api/v1/init/check")
    if (checkResp.status === 200) {
        let checkRespJson = (await checkResp.json()).data
        if (checkRespJson.userCount === 0) {
            alertDialog({
                title: "站点初始化",
                type: "form",
                tip: "设置管理员账户(请牢记密码)",
                items: [
                    {
                        label: "邮箱",
                        name: "email",
                        required: true,
                        regex: "^([a-zA-Z0-9_\.\-])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$",
                        regexerror: "邮箱格式错误"
                    },
                    {
                        label: "密码",
                        name: "password",
                        required: true
                    },
                ]
            }, async (data) => {
                let setRootResp = await fetch("/api/v1/init/root", {
                    method: "PUT",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(data)
                })
                let body = await setRootResp.json()
                if (setRootResp.status === 200) {
                    if (body.code == 200) {
                        setTimeout(() => {
                            window.location.href = home
                        }, 500)
                    }
                } else {
                    return body.data
                }
            }, () => {
                console.log("取消初始化")
            })
        } else {
            successcallback()
        }
    }
}

function initLoadMoreBtn() {
    let isLoadding = false
    let curpage = 1
    let loadMoreBtn = $("#loadmore")
    let postsContainer = $("#posts-container")
    loadMoreBtn.on("click", async () => {
        if (!isLoadding) {
            isLoadding = true
            curpage = curpage + 1
            let newcontentResp = await fetch(`${pageurl}/${curpage}`)
            if (newcontentResp.status === 200) {
                let resphtml = await newcontentResp.text()
                if (resphtml.length <= 20) {
                    loadMoreBtn.text("动态已经全部加载完毕")
                } else {
                    postsContainer.append($(`<div id="page${curpage}">${resphtml}</div>`))
                    isLoadding = false
                }
            }
        }
    })
}

function initPostComponents(postid, postuserid) {
    let cib = new CommentInputBox(postid)
    let lcb = new LikeCommentBox(postid, postuserid);
    cib.setLikeCommentBox(lcb)
    lcb.setCommentInputBox(cib)
    new LikeCommentMenu(postid, lcb, cib)
}

class CommentInputBox {
    constructor(postid) {
        let postSelector = "#post" + postid
        let postContent = $(postSelector).find(".content-0")
        this.commentinputDiv = $(`<div class="inputreply" style="height:0px">
             <div class="inner-inputreply">
                <textarea class="commenttext"  placeholder="说些什么吧" maxlength="100"></textarea>
                <div class="comment-remainnum">
                    <span class="comment-remainnum-1">0</span>/100
                </div>
                <div class="otherinfo">
                    <div class="otherinfoinput">
                        <input class="email" placeholder="输入邮箱..."></input><input class="nickname" placeholder="输入昵称..."></input>
                    </div>
                    <div class="button pubbutton">
                        评论
                    </div>
                </div>
            </div>
            <div class="closerinputreply">
                <a href="javascript:void(0)" id="closebtn">收起评论框</a>
            </div>
        </div>`)

        let closebtn = this.commentinputDiv.find("#closebtn")
        closebtn.click(() => {
            this.commentinputDiv.css("height", "0px")
        })

        let emailregex = new RegExp("^([a-zA-Z0-9_\.\-])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$");

        let emailinput = this.commentinputDiv.find(".email")
        let nameinput = this.commentinputDiv.find(".nickname")
        let pubbutton = this.commentinputDiv.find(".pubbutton")
        let commenttext = this.commentinputDiv.find(".commenttext")
        let commentremainnum = this.commentinputDiv.find(".comment-remainnum-1")

        this.commenttext = commenttext
        this.emailinput = emailinput
        this.nameinput = nameinput

        emailinput.on("focus", async () => {
            emailinput.css("color", "")
        })

        emailinput.on("blur", async () => {
            let email = emailinput.val()
            if (emailregex.test(email)) {
                let userinfoResp = await fetch(`/api/v1/user/${email}`)
                if (userinfoResp.status === 200) {
                    let username = (await userinfoResp.json()).data.username
                    nameinput.val(username)
                    nameinput.attr("readonly", "readonly")
                } else {
                    nameinput.removeAttr("readonly")
                }
            } else {
                emailinput.css("color", "red")
            }
        })

        pubbutton.on("click", async () => {
            let comment = commenttext.val().trim()
            let email = emailinput.val().trim()
            let username = nameinput.val().trim()

            if (!loginuser) {
                if (!emailregex.test(email)) {
                    alert("邮箱格式不正确")
                    return
                }
                if (username === "") {
                    alert("请输入用户名")
                    return
                }
            }

            if (comment === "") {
                alert("请输入评论内容")
                return
            }

            let data = {
                postId: postid,
                content: comment
            }

            let replytoStr = commenttext.attr("replyto")
            if (replytoStr) {
                let replyto = JSON.parse(replytoStr)
                data["replyto"] = {
                    uid: replyto.uid,
                    cid: replyto.cid
                }
            }

            let pubcommentResp = await fetch(`/api/v1/comment?email=${email}&username=${username}`, {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(data)
            })

            let body = await pubcommentResp.json()
            if (pubcommentResp.status === 200) {
                if (body.data.comment.isApproved) {
                    this.lcb.addComment(body.data)
                    toast("评论发布成功")
                } else {
                    toast("评论将在审核通过后展示")
                }

                closebtn.click()
                commenttext.val("")
                commentremainnum.text("0")

            } else {
                toast("发布评论出错,原因是:" + body.data)
            }


        })


        commenttext.on("input", () => {
            let comment = commenttext.val()
            commentremainnum.text(comment.length)
        })

        postContent.after(this.commentinputDiv)
    }

    show(reply) {
        if (!reply) {
            this.commenttext.removeAttr("replyto")
        } else {
            this.commenttext.attr("replyto", JSON.stringify(reply))
            this.commenttext.attr("placeholder", "回复给" + reply.username)
        }

        if (loginuser) {
            this.nameinput.css("display", "none")
            this.emailinput.css("display", "none")
        }
        this.commentinputDiv.css("height", "170px")
    }

    setLikeCommentBox(lcb) {
        this.lcb = lcb
    }

}

class LikeCommentMenu {
    constructor(postid, lcb, cib) {
        let postSelector = "#post" + postid
        let postContent = $(postSelector).find(".content-0")

        let buttons = $(`<div class="buttons"><div class="opbuttons">
        <div class="inner-opbuttons">
            <div class="button" id="like">
                <i class="iconfont icon-like1"></i>
            <span class="liketext">赞</span>
            </div><div class="button" id="reply">
                <i class="iconfont icon-comment-dots"></i>
            <span>评论</span></div>
        </div>
    </div>
    <div class="dotbutton">··</div></div>`)

        let dotbutton = buttons.find(".dotbutton")
        let opbuttons = buttons.find(".opbuttons")
        let likebutton = buttons.find("#like")
        let replybutton = buttons.find("#reply")

        let autoCloseTimer;

        let closeOpbuttons = () => {
            clearTimeout(autoCloseTimer)
            opbuttons.css("width", "0px")
        }

        opbuttons.on("mouseenter", () => {
            clearTimeout(autoCloseTimer)
        })

        opbuttons.on("mouseleave", () => {
            clearTimeout(autoCloseTimer)
            autoCloseTimer = setTimeout(() => {
                if (opbuttons.css("width") !== "0px") {
                    opbuttons.css("width", "0px")
                }
            }, 2000)
        })

        dotbutton.click(() => {
            if (opbuttons.css("width") == "" || opbuttons.css("width") === "0px") {

                if (lcb.isLiked) {
                    // 已经点过赞了，修改为红色
                    liketext.text("取消")
                    liketext.css("color", "#F56C6C")
                    likeicon.css("color", "#F56C6C")
                    likeicon.removeClass("icon-like1")
                    likeicon.addClass("icon-like")
                } else {
                    liketext.text("赞")
                    liketext.css("color", "#ffffff")
                    likeicon.css("color", "#ffffff")
                    likeicon.removeClass("icon-like")
                    likeicon.addClass("icon-like1")
                }

                opbuttons.css("width", "152px")
                clearTimeout(autoCloseTimer)
                autoCloseTimer = setTimeout(() => {
                    if (opbuttons.css("width") !== "0px") {
                        opbuttons.css("width", "0px")
                    }
                }, 3000)
            } else {
                closeOpbuttons()
            }
        })

        let likeicon = likebutton.find("i")
        let liketext = likebutton.find(".liketext")

        likebutton.click(async () => {

            closeOpbuttons()

            if (!lcb.isLiked) {
                let likeresp = await fetch(`/api/v1/like/post/${postid}`, {
                    method: "PUT"
                })
                if (likeresp.status === 200) {
                    lcb.isLiked = true
                    // 点赞成功
                    if (loginuser) {
                        lcb.addLike({
                            user: loginuser
                        })
                    } else {
                        lcb.addVisitorLike(1)
                    }

                } else {
                    let error = (await likeresp.json()).data
                    toast("点赞失败，原因是:" + error)
                }
            } else {
                let likeresp = await fetch(`/api/v1/like/post/${postid}`, {
                    method: "DELETE"
                })
                if (likeresp.status === 200) {
                    lcb.isLiked = false
                    // 取消点赞成功
                    if (loginuser) {
                        lcb.addLike({
                            user: loginuser
                        }, true)
                    } else {
                        lcb.addVisitorLike(-1)
                    }

                } else {
                    let error = (await likeresp.json()).data
                    toast("取消点赞失败，原因是:" + error)
                }
            }

        })

        replybutton.click(() => {
            closeOpbuttons()
            cib.show()
        })

        postContent.append(buttons)

    }
}

class LikeCommentBox {
    constructor(postid, postuserid) {
        this.init(postid, postuserid)
    }

    async init(postid, postuserid) {
        let postSelector = "#post" + postid
        this.postContent = $(postSelector).find(".content")
        let commentsResp = await fetch(`/api/v1/comment/post/${postid}?loadAll=${isDetail}`)
        if (commentsResp.status === 200) {
            let data = (await commentsResp.json()).data
            this.likes = data.likes.likeDetails || []
            this.visitorLikeCount = 0
            this.comments = data.comments || []
            this.isLiked = data.likes.isLiked
            if (this.likes.length !== 0 || this.comments.length !== 0) {
                this.createCommentBox()
            }

            if (this.likes.length != 0) {
                for (let i = 0; i < this.likes.length; i++) {
                    this.addLike(this.likes[i], false)
                }
            }

            if (data.likes.visitorLikeCount > 0) {
                this.addVisitorLike(data.likes.visitorLikeCount)
            }

            if (this.comments.length != 0) {
                if (isDetail) {
                    for (let i = 0; i < this.comments.length; i++) {
                        this.addComment(this.comments[i], postuserid)
                    }
                } else {
                    for (let i = 0; i < this.comments.length && i < 5; i++) {
                        this.addComment(this.comments[i], postuserid, this.comments.length > 5 && i === 4)
                    }
                }


            }

            this.createLine()
        }
    }

    createCommentBox() {
        this.commentBox = $(`<div class="content-1"></div>`)
        this.postContent.append(this.commentBox)
    }

    createLikeBox() {
        if (!this.commentBox) {
            this.createCommentBox()
        }
        if (!this.likesBox) {
            this.likesBox = $(`<div class="likes">
                <i class="iconfont icon-like1"></i>
            </div>`)
            this.commentBox.prepend(this.likesBox)

        }
    }

    createReplysBox() {
        if (!this.commentBox) {
            this.createCommentBox()
        }
        if (!this.replysBox) {
            this.replysBox = $(`<div class="replys"></div>`)
            this.commentBox.append(this.replysBox)
        }
    }

    createLine() {
        if ((this.likes.length > 0 || this.visitorLikeCount > 0) && this.comments.length > 0 && !this.hasline) {
            this.hasline = true
            let linediv = $(`<div class="line"></div>`)
            this.likesBox.after(linediv);
        }
    }

    addVisitorLike(count) {
        this.visitorLikeCount = this.visitorLikeCount + count

        this.createLikeBox()
        this.createLine()

        if (!this.countSpan) {
            this.countdiv = $(`<a class="count" href="javascript:void(0)"><span1 class="likecount">0</span1>个访客</a>`)
            this.countSpan = this.countdiv.find(".likecount")
            this.likesBox.append(this.countdiv)
        }


        this.countSpan.text(this.visitorLikeCount)

    }

    addLike(like, isDelete) {
        this.createLikeBox()
        this.createLine()
        if (isDelete) {
            let likeitems = this.likesBox.find("a")
            if (likeitems.length > 0) {
                for (let i = 0; i < likeitems.length; i++) {
                    let name = likeitems[i].text
                    if (name === (like.user.user.username + ",")) {
                        likeitems[i].remove()
                        break
                    }
                }
            }

            likeitems = this.likesBox.find("a")
            if (likeitems.length === 0) {
                this.likesBox.remove()
                this.likesBox = undefined
            }

        } else {
            let likeitem = $(`<a href="/user/${like.user.user.ID}">${like.user.user.username},</a>`)
            if (this.countdiv) {
                this.countdiv.before(likeitem)
            } else {
                this.likesBox.append(likeitem)
            }
        }
    }

    addComment(comment, postuserid, hasMore) {
        this.createReplysBox()
        this.createLine()

        let isApproved = comment.comment.isApproved
        let innerhtml = `<span class="name">${escapeHTML(comment.user.user.username)}</span>`
        if (comment.replyto) {
            innerhtml += `<span class="text"> 回复 </span><span class="name">${escapeHTML(comment.replyto.user.username)}</span>`
        }
        if (hasMore && !isDetail) {
            innerhtml += `: <span class="text commentcontent">${escapeHTML(comment.comment.content)}</span>&nbsp;&nbsp;<a href="/post/${comment.comment.postId}" style="color:#409EFF;">查看更多评论</a>`
        } else {
            innerhtml += `: <span class="text commentcontent">${escapeHTML(comment.comment.content)}</span>`
        }


        let cuser = currentUser()

        // 是不是自己动态下的评论或者自己是不是超级管理员?
        let isCanDelete = false
        if (cuser) {
            isCanDelete = cuser.role === "admin" || postuserid === cuser.uid || comment.comment.userId === cuser.uid
            if (isCanDelete) {
                innerhtml += `&nbsp;&nbsp;<a href="javascript:void(0)" class="delete"><i class="iconfont icon-delete1"></i></a>`
            }
        }


        if (!isApproved) {
            innerhtml += `&nbsp;&nbsp;<a href="javascript:void(0)" class="support" >批准此评论</a>`
        }

        let replydiv = $(`<div class="reply">${innerhtml}</div>`)
        if (isApproved) {
            replydiv.find(".commentcontent").click(() => {
                console.log(comment)
                this.cib.show({
                    uid: comment.user.user.ID,
                    cid: comment.comment.ID,
                    username: comment.user.user.username,
                })
            })
        }
        if (isCanDelete) {
            let deleteBtn = replydiv.find(".delete")
            deleteBtn.on("click", () => {
                warn("确定要删除此条评论吗？此操作不可逆", async () => {
                    let deleteResp = await fetch(`/api/v1/comment/${comment.comment.ID}`, {
                        method: "DELETE"
                    })
                    if (deleteResp.status === 200) {
                        replydiv.remove()
                    } else {
                        toast("操作失败")
                    }
                })
            })

        }
        if (!isApproved) {
            let supportBtn = replydiv.find(".support")
            supportBtn.on("click", () => {
                confirm(`确定要批准这条评论吗？评论内容是: ${escapeHTML(comment.comment.content)}`, async () => {
                    let approveResp = await fetch(`/api/v1/comment/approve/${comment.comment.ID}`, {
                        method: "POST"
                    })
                    if (approveResp.status === 200) {
                        supportBtn.remove()
                        deleteBtn.remove()
                    } else {
                        toast("操作失败")
                    }

                })
            })

        }

        this.replysBox.append(replydiv)
    }

    setCommentInputBox(cib) {
        this.cib = cib
    }

}

async function initBanner() {

    $("#camerabtn").click(() => {

        alertDialog({
            title: "请选择发布类型",
            type: "options",
            items: [
                {
                    label: "普通图文",
                    value: "text",
                }, {
                    label: "音乐",
                    value: "music",
                }, {
                    label: "卡片",
                    value: "card",
                },
                {
                    label: "外站音视频",
                    value: "external",
                },
            ]
        }, async (data) => {
            resetEditPage(data.value)
            $("#publishpostPage").css("display", "block")
            setTimeout(() => {
                $("#publishpostPage").css("top", "0px")
                $("#posts-container").css("display", "none")
            }, 10)
        });


    })

    $("#settingbtn").click(() => {
        resetSettingPage()
        $("#settingpage").css("display", "block")
        setTimeout(() => {
            $("#settingpage").css("top", "0px")
            $("#posts-container").css("display", "none")
        }, 10)
    })

    $("#homeicon").click(() => {
        window.location.href = "/"
    })

    let login = async () => {
        alertDialog({
            title: "登录",
            type: "form",
            tip: `登录到[${sitename}]`,
            items: [
                {
                    label: "邮箱",
                    name: "email",
                    required: true,
                    regex: "^([a-zA-Z0-9_\.\-])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$",
                    regexerror: "邮箱格式错误"
                },
                {
                    label: "密码",
                    name: "password",
                    required: true
                },
            ]
        }, async (data) => {
            let loginResp = await fetch("/api/v1/login", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(data)
            })
            let body = await loginResp.json()

            if (loginResp.status === 200) {
                if (body.code == 200) {
                    let token = body.data
                    addUser(token)
                    setTimeout(() => {
                        window.location.href = home
                    }, 100)
                }
            } else {
                return body.data
            }
        }, () => {
            console.log("用户取消了登录")
        })
    }

    $("#loginbtn").click(login)


    $("#switchAccountbtn").click(() => {
        let users = listUsers()
        let items = []
        let loginuserEmail = loginuser.user.email
        for (let i = 0; i < users.length; i++) {
            items.push({
                label: users[i].username + (loginuserEmail === users[i].email ? '(当前用户)' : ''),
                value: users[i].email
            })
        }
        items.push({
            label: "添加新的账户",
            value: "addNewAccount"
        })
        alertDialog({
            title: "切换账号",
            type: "options",
            items: items
        }, async (data) => {
            if (data.value !== "addNewAccount") {
                let email = data.value
                if (email !== loginuserEmail) {
                    changeUser(email)
                    setTimeout(() => {
                        window.location.href = home
                    }, 100)
                }

            } else {
                login()
            }

        });
    })

    $("#logoutbtn").click(async () => {
        confirm(`确定要退出账号[${loginuser.user.username}]吗?`, () => {
            logoutUser(loginuser.user.email)
            setTimeout(() => {
                window.location.href = home
            }, 100)
        })

    })


}

async function editPost(postid) {
    let postResp = await fetch(`/api/v1/post/${postid}`)
    if (postResp.status === 200) {
        let postjson = await postResp.json()
        if (postjson.code === 200) {
            let post = postjson.data.post
            resetEditPage("text", post)
            $("#publishpostPage").css("display", "block")
            setTimeout(() => {
                $("#publishpostPage").css("top", "0px")
                $("#posts-container").css("display", "none")
            }, 10)
        }

    }
}

function playVideo(parentid) {
    imageViewer.initViewer(parentid, 0, "video")
}

function showPlayButton(imgElement, parentid) {
    const imgWidth = imgElement.clientWidth;
    const parentElement = imgElement.parentElement;

    let playbutton = $(`<div class="playbutton">
        <i class="iconfont icon-play-filling"></i>
        </div>`)

    playbutton.on("click", () => {
        playVideo(parentid)
    })

    playbutton.css("left", (imgWidth / 2) + "px")
    $(parentElement).append(playbutton)
}



function escapeHTML(input) {
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(input));
    return div.innerHTML;
}




class Account {

    constructor(container, account) {

        let userdiv = $(`<li class="row">
        <div class="avatar" id="avatarset">
        <img src="${account.avatar}" />
    </div>
    <div class="a-right">
        <div class="accountinfo">
            <div class="a-user">${account.username}</div>
            <div class="a-email">${account.email}</div>
        </div>
        <div class="a-btn">
           
        </div>
    </div>
    
        </li>`)

        let btns = userdiv.find(".a-btn")
        if (account.role === "user") {
            let enablebutton = $(`<a href="javascript:void(0)">${account.enable ? "禁用" : "启用"}</a>`)
            if (account.enable) {
                enablebutton.addClass("disable")
            }

            enablebutton.click(() => {
                let newenable = !account.enable;
                confirm(`确定将[${account.username}]设置为${newenable ? "启用" : "禁用"}状态吗`, async () => {
                    let setResp = await fetch(`/api/v1/admin/user/enable/${account.id}`, {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/json",
                        },
                        body: JSON.stringify({
                            enable: newenable
                        })
                    })
                    if (setResp.status === 200) {
                        enablebutton.text(newenable ? "禁用" : "启用")
                        account.enable = newenable
                        if (newenable) {
                            enablebutton.addClass("disable")
                        } else {
                            enablebutton.removeClass("disable")
                        }
                    } else {
                        let err = (await setResp.json()).data
                        toast("修改失败，原因是:" + err)
                    }
                })
            })

            btns.append(enablebutton)
        }

        let resetPwdButton = $(`<a href="javascript:void(0)">重置密码</a>`)
        resetPwdButton.click(async () => {
            alertDialog({
                title: "重置用户密码",
                type: "form",
                items: [
                    {
                        label: "密码",
                        name: "password",
                        required: true
                    },
                ]
            }, async (data) => {
                let newpassword = {
                    password: data.password
                }
                let setResp = await fetch(`/api/v1/admin/user/password/${account.id}`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(newpassword)
                })

                if (setResp.status === 200) {
                    toast("密码重置成功")
                } else {
                    let err = (await setResp.json()).data
                    toast("密码重置失败，原因是:" + err)
                }
            }, () => {
                console.log("用户取消了操作")
            })
        })

        btns.append(resetPwdButton)

        container.append(userdiv)
    }
}

class AccountList {
    constructor() {
        this.init = false
    }

    async tryInit() {
        if (!this.init) {
            this.accountlist = $("#accountlist")
            this.accountlist.html("")
            let usersResp = await fetch("/api/v1/admin/user/all")
            if (usersResp.status === 200) {
                let users = (await usersResp.json()).data
                for (let i = 0; i < users.length; i++) {
                    new Account(this.accountlist, users[i])
                }
            }
            this.init = true
        }
    }

    addNewAccount(newaccount) {
        if (this.init) {
            new Account(this.accountlist, newaccount)
        }

    }

}

let accountlist = new AccountList()
let avatarsetUploader;
let bannersetUploader;

function initSettingPage() {

    let exitsettingbtn = $("#operator").find(".exitsetting")
    let savesettingbtn = $("#operator").find(".savesetting")

    let postsContainer = $("#posts-container")
    let settingpage = $("#settingpage")

    exitsettingbtn.on("click", () => {
        postsContainer.css("display", "block")
        settingpage.css("top", "500px")
        setTimeout(() => {
            settingpage.css("display", "none")
        }, 250)
    })


    let nicknameset = $("#nicknameset")
    let signset = $("#signset")

    avatarsetUploader = new Uploader($("#avatarset"), {
        types: ["image"]
    })

    bannersetUploader = new Uploader($("#bannerset"), {
        types: ["image"]
    })

    savesettingbtn.on("click", async () => {
        let coverurl = bannersetUploader.getUrl()
        let avatarseturl = avatarsetUploader.getUrl()

        if (!coverurl || !avatarseturl) {
            toast("图片或者封面上传失败")
            return
        }

        let updateuser = {
            user: {
                username: nicknameset.val(),
            },
            profile: {
                cover: coverurl.url,
                avatar: avatarseturl.url,
                signature: signset.text()
            }
        }


        let updateResp = await fetch("/api/v1/user", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(updateuser)
        })

        if (updateResp.status === 200) {
            alert("修改成功,刷新页面即可看到修改效果")
        } else {
            let respjson = await updateResp.json()
            alert("修改失败,原因是" + respjson.data)
        }


    })


    let newuserbtn = $("#newuser").find("a")
    newuserbtn.on("click", () => {
        alertDialog({
            title: "添加新用户",
            type: "form",
            tip: "其它信息需要用户登录后自行设置",
            items: [
                {
                    label: "邮箱",
                    name: "email",
                    required: true,
                    regex: "^([a-zA-Z0-9_\.\-])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$",
                    regexerror: "邮箱格式错误"
                },
                {
                    label: "密码",
                    name: "password",
                    required: true
                },
            ]
        }, async (data) => {
            let newuser = {
                email: data.email,
                password: data.password
            }
            let addNewUserResp = await fetch("/api/v1/admin/user", {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(newuser)
            })

            if (addNewUserResp.status === 200) {
                let respjson = await addNewUserResp.json()
                let newuser = respjson.data
                accountlist.addNewAccount({
                    avatar: newuser.profile.avatar,
                    email: newuser.user.email,
                    enable: newuser.user.enable,
                    id: newuser.user.ID,
                    role: newuser.user.role,
                    username: newuser.user.username
                })
            } else {
                let respjson = await addNewUserResp.json()
                return respjson.data
            }
        }, () => {
            console.log("用户取消了操作")
        })
    })

    let accountset = $("#settingpage").find(".accountset")
    accountset.css("display", "none")

}

async function resetSettingPage() {

    let nicknameset = $("#nicknameset")
    let signset = $("#signset")
    let emailset = $("#emailset")


    let userinfoResp = await fetch("/api/v1/user")

    if (userinfoResp.status === 200) {

        let userinfo = await userinfoResp.json()
        let email = userinfo.data.user.email
        let username = userinfo.data.user.username
        let signature = userinfo.data.profile.signature
        let cover = userinfo.data.profile.cover
        let avatar = userinfo.data.profile.avatar
        let role = userinfo.data.user.role

        emailset.text(email)
        nicknameset.val(username)
        signset.html(signature)
        avatarsetUploader.setInitUrl({
            type: "image",
            url: avatar
        })
        bannersetUploader.setInitUrl({
            type: "image",
            url: cover
        })


        if (role === "admin") {
            await accountlist.tryInit();
            let accountset = $("#settingpage").find(".accountset")
            accountset.css("display", "block")

        }

    }
}


class ExternalMediaBox {
    constructor() {
        this.init()
        this.reset()
    }

    async init() {
        let paltformbuttons = $("#publishpostPage").find(".paltformbuttons")
        let preview = $("#publishpostPage").find(".preview")

        let configResp = await fetch("/api/v1/post/externalmedia/config")
        if (configResp.status === 200) {
            let config = (await configResp.json()).data

            Object.keys(config).forEach(key => {
                let videoConfigs = config[key]
                for (let i = 0; i < videoConfigs.length; i++) {
                    let videoConfig = videoConfigs[i]
                    let button = $(`<button>${videoConfig.displayname}</button>`)
                    let items = []
                    for (let j = 0; j < videoConfig.props.length; j++) {
                        let prop = videoConfig.props[j]
                        let item = {
                            label: prop.displayname,
                            name: prop.name,
                            required: true
                        }
                        items.push(item)
                    }
                    button.click(() => {
                        paltformbuttons.find("button").removeClass("click")
                        button.addClass("click")
                        alertDialog({
                            title: `添加${videoConfig.displayname}`,
                            type: "form",
                            items: items
                        }, async (data) => {
                            this.config = {
                                type: key,
                                name: videoConfig.name,
                                config: data
                            }
                            let previewhtmlResp = await fetch(`/api/v1/post/externalmedia/preview?config=${JSON.stringify(this.config)}`)
                            if (previewhtmlResp.status === 200) {
                                let previewhtml = (await previewhtmlResp.json()).data
                                preview.html(previewhtml)
                            }
                        })
                    })
                    paltformbuttons.append(button)
                }
            })


        }
    }

    reset() {
        this.config = undefined
        let paltformbuttons = $(".paltform").find(".paltformbuttons")
        paltformbuttons.find("button").removeClass("click")
        let preview = $(".paltform").find(".preview")
        preview.text("预览区域")
    }

    getData() {
        return this.config
    }
}

var externalMediaBox
var cardBox;

let musicCoverUploader;

function isSelectionInsideDiv(selection, div) {
    if (!selection.rangeCount) return false;

    const range = selection.getRangeAt(0);
    return div.contains(range.startContainer) && div.contains(range.endContainer);
}

function initEditPage() {

    let cancelpublishpostPageBtn = $("#cancelpublishpostPageBtn")
    let publishpostPageBtn = $("#publishpostPageBtn")
    let postsContainer = $("#posts-container")
    let publishpostPage = $("#publishpostPage")
    let locationinputbtn = $("#locationinputbtn")
    let addressinputbox = $("#addressinputbox")
    let visibilitySwitch = $(".visibility").find("input")
    let uploadMusicBtn = $("#uploadMusic")
    let uploadMusicBar = $(".uploadMusicBar")
    let musicinfo = $(".musicinfo")
    let inputs = musicinfo.find("input")
    let titleInput = inputs[0]
    let artistInput = inputs[1]
    let albumInput = inputs[2]
    let addLinkBtn = $(".addLinkBtn")
    let editpostText = $("#editpost-text")

    cancelpublishpostPageBtn.on("click", () => {
        postsContainer.css("display", "block")
        publishpostPage.css("top", "500px")
        setTimeout(() => {
            publishpostPage.css("display", "none")
            cardBox.clear()
        }, 300)
    })

    locationinputbtn.on("click", () => {
        if (addressinputbox.css("height") !== "50px") {
            addressinputbox.css("height", "50px")
        } else {
            addressinputbox.css("height", "0px")
        }
    })

    addLinkBtn.on("click", () => {
        let selection = window.getSelection();
        let range = undefined
        if (isSelectionInsideDiv(selection, editpostText[0])) {
            if (selection.rangeCount > 0) {
                range = selection.getRangeAt(0);
            }
        }
        alertDialog({
            title: "插入链接",
            type: "form",
            items: [
                {
                    label: "标题",
                    name: "title",
                    required: true
                },
                {
                    label: "URL",
                    name: "url",
                    required: true,
                    regex: "^(http:\/\/|https:\/\/)",
                    regexerror: "请输入合法的url"
                },
            ]
        }, async (data) => {
            let title = data.title
            let url = data.url
            let innsettext = `[url href="${url}"]${title}[/url]`

            if (range) {
                range.deleteContents();
                const textNode = document.createTextNode(innsettext);
                range.insertNode(textNode);

                range.setStartAfter(textNode);
                range.setEndAfter(textNode);
                selection.removeAllRanges();
                selection.addRange(range);
            } else {
                const textNode = document.createTextNode(innsettext);
                editpostText.append(textNode);
            }


        }, () => {

        })
    })

    externalMediaBox = new ExternalMediaBox();
    cardBox = new CardBox();


    editpostText.on("blur", () => {
        cardBox.setContent(editpostText[0].innerText)
    })

    let addressinputs = addressinputbox.find("input")
    publishpostPageBtn.on("click", async () => {
        if ($("#publishpostPageBtn").hasClass("disabled")) {
            return
        }

        let post = {
            content: editpostText[0].innerText,
        }

        if (postid) {
            post.ID = postid
        }

        let address = "";
        for (let i = 0; i < addressinputs.length; i++) {
            let v = addressinputs[i].value
            if (v === "") {
                break
            } else {
                if (i === 0) {
                    address = v
                } else {
                    address = address + " · " + v
                }
            }
        }

        if (address !== "") {
            Cookies.set(`address_${loginuser.user.email}`, address, { expires: 365 })
        }

        post.address = address

        post.medias = []
        for (let i = 0; i < editUploaders.length; i++) {
            let editUploader = editUploaders[i]
            let uploadResp = editUploader.getUrl()
            if (uploadResp) {
                post.medias.push({
                    type: uploadResp.type,
                    url: uploadResp.url,
                    thumbnail: uploadResp.cover
                })
            }
        }

        if (post.medias.length === 0) {
            // 检查是不是上传的音乐
            let musicurl = uploadMusicBtn.attr("url")
            if (musicurl) {
                let duration = parseInt(musicinfo.attr("duration"))
                let media = {
                    type: "music",
                    url: musicurl,
                    title: titleInput.value,
                    artist: artistInput.value,
                    ablum: albumInput.value,
                    duration: duration,
                }

                let coverurl = musicCoverUploader.getUrl();
                if (coverurl) {
                    media.thumbnail = coverurl.url
                }

                post.medias.push(media)
            }
        }


        post.externalMedia = undefined
        let externalMedia = externalMediaBox.getData()
        if (externalMedia) {
            post.externalMedia = externalMedia
        }

        post.card = undefined
        let carddata = cardBox.getData()
        if (carddata) {
            if (cardBox.checkData()) {
                post.card = carddata
            } else {
                alert("请填写所有必填项")
                return
            }
        }


        post.visibility = visibilitySwitch.prop('checked') ? "private" : "public";

        if (post.content === "" && post.url === undefined
            && (post.medias === undefined || post.medias.length === 0) && post.externalMedia === undefined && post.card == undefined) {
            alert("无法发布空的内容")
        } else {
            let method = postid === undefined ? "PUT" : "POST"
            let postResp = await fetch("/api/v1/post", {
                method: method,
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(post)
            })

            if (postResp.status === 200) {
                cancelpublishpostPageBtn.click()
                externalMediaBox.reset()
                setTimeout(() => {
                    window.location.href = home
                }, 500)
            } else {
                let error = (await postResp.json()).data
                warn(error)
            }

        }
    })

    musicCoverUploader = new Uploader($("#musiccover"), {
        types: ["image"]
    })

    let uploading = false

    uploadMusicBtn.on("click", async () => {
        if (uploading) {
            return
        }
        uploading = true
        let input = document.createElement("input")
        input.type = "file"
        input.accept = "audio/mp3"
        input.addEventListener('cancel', () => {
            uploading = false
        })
        input.addEventListener("change", event => {
            let files = event.target.files
            if (files.length > 0) {
                let file = files[0]
                if (file.type.indexOf("audio") === 0) {
                    let closeModal = modalDialog("请稍等，正在解析音乐....")
                    uploadMusicBar.css("background", "#409EFF")
                    let getAudioDuration = (callback) => {
                        callback = callback || function () { }
                        let audio = new Audio()
                        audio.addEventListener('loadedmetadata', () => {
                            console.log(audio.duration)
                            callback(audio.duration)
                        })
                        audio.src = URL.createObjectURL(file)
                    }
                    let uploadMusic = () => {
                        closeModal()
                        // 上传过程中禁止提交
                        $("#publishpostPageBtn").addClass("disabled")
                        getAudioDuration(duration => {
                            musicinfo.attr("duration", Math.floor(duration) + "")
                            uploadFile(file, resp => {
                                uploadMusicBtn.attr("url", resp.url)
                                uploadMusicBar.css("background", "#07BF5F")
                                $("#publishpostPageBtn").removeClass("disabled")
                                uploading = false
                            }, () => {
                                toast("上传音乐失败")
                                uploadMusicBar.css("background", "#FA5251")
                                $("#publishpostPageBtn").removeClass("disabled")
                                uploading = false
                            }, progress => {
                                uploadMusicBar.css("width", progress + "%")
                            }
                            )
                        })


                    }
                    jsmediatags.read(file, {
                        onSuccess: (tag) => {
                            if (tag.tags) {

                                let tags = tag.tags
                                let title = tags.title
                                let artist = tags.artist
                                let album = tags.album
                                let picture = tags.picture

                                if (title) {
                                    titleInput.value = title
                                } else {
                                    titleInput.value = file.name
                                }
                                if (artist) {
                                    artistInput.value = artist
                                } else {
                                    artistInput.value = ""
                                }
                                if (album) {
                                    albumInput.value = album
                                } else {
                                    albumInput.value = ""
                                }

                                if (picture && picture.data) {
                                    const { data, format } = picture;
                                    const byteArray = new Uint8Array(data);
                                    const blob = new Blob([byteArray], { type: format });
                                    uploadFile(blob, resp => {
                                        musicCoverUploader.setInitUrl({
                                            type: "image",
                                            url: resp.url
                                        })

                                        uploadMusic()
                                    })

                                } else {
                                    musicCoverUploader.reset();
                                    uploadMusic()
                                }
                            }
                        },
                        onError: (error) => {
                            titleInput.value = file.name
                            artistInput.value = ""
                            albumInput.value = ""
                            musicCoverUploader.reset();
                            uploadMusic()
                        }
                    });

                }
            }
        })
        input.click()

    })

    let posbtn = addressinputbox.find(".getPosBtn")
    posbtn.on("click", async () => {
        let iplocResp = await fetch(`/api/v1/lbs/mypos`)
        if (iplocResp.status === 200) {
            let locdata = await iplocResp.json()
            if (locdata.code === 200) {
                for (let i = 0; i < addressinputs.length; i++) {
                    addressinputs[i].value = ""
                }
                let myaddress = locdata.data.address
                for (let i = 0; i < myaddress.length; i++) {
                    addressinputs[i].value = myaddress[i]
                }

                return
            }
        }

        toast("定位失败，可以手动填写位置")

    })

}

let editUploaders = []
let postid = undefined
function resetEditPage(posttype, post) {
    $(".post-a").css("display", "none")
    if (posttype === "external") {
        $(".posttype_extenral").css("display", "block")
    } else if (posttype === "music") {
        $(".posttype_music").css("display", "block")
    } else if (posttype === "card") {
        $(".posttype_card").css("display", "block")
        cardBox.reset()
    } else {
        $(".posttype_image").css("display", "block")
    }

    if (post) {
        postid = post.ID
    } else {
        postid = undefined
    }


    let visibilitySwitch = $(".visibility").find("input")
    visibilitySwitch.prop("checked", false)

    if(post && post.visibility === "private"){
        visibilitySwitch.prop("checked", true)
    }

    externalMediaBox.reset()

    let editpostText = $("#editpost-text")
    let pics = $("#postpics")
    let addressinputbox = $("#addressinputbox")


    // 文本框
    editpostText.html("")
    if (post) {
        editpostText.html(post.content)
    }

    // 地址
    addressinputbox.css("height", "0px")
    let addressinput = addressinputbox.find("input")
    addressinput.val("")
    let initaddress = ""
    if (post) {
        initaddress = post.address
    } else {
        initaddress = Cookies.get(`address_${loginuser.user.email}`)
    }
    if (initaddress) {
        let saveaddresses = initaddress.split("·")
        for (let j = 0; j < saveaddresses.length && j < 4; j++) {
            addressinput[j].value = saveaddresses[j].trim()
        }
    }



    let musicinfo = $(".musicinfo")
    musicinfo.removeAttr("duration")
    let inputs = musicinfo.find("input")
    inputs[0].value = ""
    inputs[1].value = ""
    inputs[2].value = ""
    $(".uploadMusicBar").css("width", "0px")
    $(".uploadMusicBar").css("background", "#409EFF")
    $("#uploadMusic").removeAttr("url")
    musicCoverUploader.reset()

    for (let i = 0; i < editUploaders.length; i++) {
        editUploaders[i].delete()
    }
    editUploaders = []
    pics.html("")
    let maxUploaders = 9
    let uploadSuccessCallback = (initMedia) => {
        if (editUploaders.length < maxUploaders) {
            if (editUploaders.length > 0) {
                let firstEditUploader = editUploaders[0]
                let firstUploadMedia = firstEditUploader.getUrl()
                if (firstUploadMedia.type === "video") {
                    return
                }
            }
            let uploaderDiv = $(`<div class="uploader"></div>`)
            let types = ["image"]
            if (editUploaders.length === 0) {
                types = ["image", "video"]
            }
            let uploader = new Uploader(uploaderDiv, {
                types: types,
                cover: true,
                allowDelete: true,
                callback: uploadResp => {
                    if (uploadResp.type === "image") {
                        uploadSuccessCallback()
                    }

                },
                deleteCallback: () => {
                    let findindex = editUploaders.findIndex(x => x === uploader)
                    if (findindex !== -1) {
                        editUploaders.splice(findindex, 1)
                    }
                    let hasUploadedNum = editUploaders.filter(x => x.getUrl() !== undefined).length;
                    if (hasUploadedNum === maxUploaders - 1) {
                        uploadSuccessCallback()
                    }

                    if (hasUploadedNum === 0) {
                        if (editUploaders.length > 0) {
                            for (let j = 0; j < editUploaders.length; j++) {
                                editUploaders[j].delete()
                            }
                        }
                        editUploaders = []
                        uploadSuccessCallback()
                    }

                }
            })
            pics.append(uploaderDiv)
            editUploaders.push(uploader)

            if (initMedia) {
                uploader.setInitUrl(initMedia)
            }

        }
    }

    if (post && post.medias && post.medias.length > 0) {
        for (let i = 0; i < post.medias.length; i++) {
            let media = post.medias[i]
            uploadSuccessCallback({
                type: media.type,
                url: media.url,
                cover: media.thumbnail,
            });
        }
    }
    uploadSuccessCallback();


}


function deletePost(pid) {
    warn(`你确定要删除这个动态吗?`, async () => {
        let deleteResp = await fetch(`/api/v1/post/${pid}`, {
            method: "DELETE"
        })
        if (deleteResp.status === 200) {
            let postid = "post" + pid
            let post = document.getElementById(postid)
            if (post) {
                post.remove()
            }
        } else {
            let resp = await deleteResp.json()
            alert("删除失败,原因是:" + resp.data)
        }
    })
}



function showImage(domid, index) {
    imageViewer.initViewer(domid, index)
}

async function setPostTop(postid, istop) {
    if (istop) {
        confirm("确定要置顶这条动态吗？置顶此条消息将会取消此用户已经置顶的动态", async () => {
            let setTopResp = await fetch(`/api/v1/post/top/${postid}`, {
                method: "POST"
            })
            if (setTopResp.status === 200) {
                setTimeout(() => {
                    window.location.href = home
                }, 500)

            }
        })
    } else {
        confirm("确定要取消置顶这条动态吗？", async () => {
            let setTopResp = await fetch(`/api/v1/post/top/${postid}`, {
                method: "Delete"
            })
            if (setTopResp.status === 200) {
                setTimeout(() => {
                    window.location.href = home
                }, 500)

            }
        })
    }

}

function scrollToTop() {
    window.scrollTo({
        top: 0,
        behavior: 'smooth'
    });
}