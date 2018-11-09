
function komment_form(in_config)
{
  var id = "komment_form_" + in_config.komment_id
  var id_comments = "komment_comments_" + in_config.komment_id

  if ($('#'+id).length == 0)
  {
    document.write('<div id="'+id+'"></div>')
  }

  $('#'+id).load("komment_form.html", null, function(){
    $('#'+id).ajaxForm({
      resetForm: true,
      data: { komment_id: in_config.komment_id },
      dataType: 'json',
      success: function() {
        alert("Thanks for the comment")
        $('#'+id_comments).load("comments_"+in_config.komment_id+".json");
        alert("Here is your comment")
      },
      error: function() {
        alert("Error!")
      }
    });
  })

}


function komment_comments(in_config)
{
  var id = "komment_comments_" + in_config.komment_id

  if ($('#'+id).length == 0)
  {
    document.write('<div id="'+id+'"/></div>')
  }

  $('#'+id).load("comments_"+in_config.komment_id+".json")
}


function komment_count(in_config)
{
  var id = "komment_form_" + in_config.komment_id
}

