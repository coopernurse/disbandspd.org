
function submitPetition() {
    var email = $("#email").val();
    if (email && email.indexOf("@") > 0 && email.indexOf(".") > 0) {
        var req = {
            "email": email
        };
        $("#formmsg").text("");
        $.post("/petition-json", req, function(resp) {
            if (resp.success) {
                $("#formmsg").text("Thank you for signing our petition! We will email you periodic updates with the progress of the campaign.");
                $("#petitionform").hide();
            } else {
                $("#formmsg").text(resp.error);
            }
        });
    } else {
        $("#formmsg").text("Please enter a valid email");
    }
}
