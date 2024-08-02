function parseJwt(token) {
    var base64Url = token.split('.')[1];
    var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    var jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    return JSON.parse(jsonPayload);
}

function addUser(token) {
    let payload = parseJwt(token)
    let username = payload.username
    let email = payload.email
    let role = payload.role
    let exp = payload.exp

    let user = {
        username: username,
        role: role,
        email: email,
        token: token,
        exp: exp,
    }

    let usersStr = localStorage.getItem("users")
    if (!usersStr) {
        usersStr = "{}"
    }

    let users = JSON.parse(usersStr)
    users[email] = user
    localStorage.setItem("users", JSON.stringify(users))

    Cookies.set("Authorization", token, { expires: 25 })
}

function deleteUser(email) {
    let usersStr = localStorage.getItem("users")
    if (!usersStr) {
        usersStr = "{}"
    }

    let users = JSON.parse(usersStr)
    delete users[email]
    localStorage.setItem("users", JSON.stringify(users))
}

function logoutUser(email){
    deleteUser(email)
    Cookies.set("Authorization", "", { expires: 0 })
    let users = listUsers()
    if(users.length > 0){
        changeUser(users[0].email)
    }
}

function listUsers() {
    let usersStr = localStorage.getItem("users")
    if (!usersStr) {
        usersStr = "{}"
    }

    let users = JSON.parse(usersStr)
    let userList = []
    Object.keys(users).forEach(key => {
        let user = users[key]
        if (user) {
            let exp = user.exp
            if ((exp * 1000 - new Date().getTime()) >= 1000 * 3600 * 24) {
                userList.push({
                    email: user.email,
                    username: user.username,
                    role: user.role,
                })
            }else{
                deleteUser(key)
            }
        }


    })
    return userList
}

function changeUser(email) {
    let usersStr = localStorage.getItem("users")
    if (!usersStr) {
        usersStr = "{}"
    }

    let users = JSON.parse(usersStr)
    let token = users[email].token
    Cookies.set("Authorization", token, { expires: 25 })
}

function currentUser(){
    let jwttoken = Cookies.get("Authorization")
    if(jwttoken){
        return parseJwt(jwttoken)
    }
}