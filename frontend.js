
function komment_form(in_config)
{
  var id = "komment_form_" + in_config.komment_id
  var id_comments = "komment_comments_" + in_config.komment_id

  if ($('#'+id).length == 0)
  {
    document.write('<div id="'+id+'"></div>')
  }

  $('#'+id).load("komment_form.html", null, function(){
    var obj = $('#'+id+' input[name="komment_id"]').get(0)
    obj.value = in_config.komment_id
    $('#'+id).ajaxForm({
      resetForm: true,
      success: function() {
        $('#'+id_comments).load("comments_"+in_config.komment_id+".txt");
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

  $('#'+id).load("comments_"+in_config.komment_id+".txt")
}


function komment_count(in_config)
{
  var id = "komment_form_" + in_config.komment_id
}

