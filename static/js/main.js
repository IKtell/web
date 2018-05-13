$(document).ready(function () {
    $(".choose").click(function () {
        location.href = "/choose"
    });
    $(".logoutButton").click(logoutAnimation);
    $("#login").click(loginAnimation);
    $(".show").click(function () {
        location.href = "/show"
    });
    $(".knowledge").click(function () {
        location.href = "/knowledge"
    });
    $(".report").click(function() {
        location.href = "/report"
    })
    function loginAnimation() {
        $(".page").addClass("bottom-out");
        $(".form-wrap").addClass("zoom-in");
        $("#loginName").val("");
        $("#loginPassword").val("");
        $(".splash").css("display", "");

        setTimeout(function () {
            $(".form-wrap").removeClass("zoom-in");
            $(".page").removeClass("bottom-out").css("display", "none");
            location.href = "/login";
        }, 2600);
    }

    function logoutAnimation() {
        $(".page").addClass("bottom-out");
        $(".form-wrap").addClass("zoom-in");
        $("#loginName").val("");
        $("#loginPassword").val("");
        $(".splash").css("display", "");

        setTimeout(function () {
            $(".form-wrap").removeClass("zoom-in");
            $(".page").removeClass("bottom-out").css("display", "none");
            $.ajax({
                url: "/logout",
                type: "GET",
                dataType: "text",
            });
            location.href = "/login";
        }, 1600);
    }
})