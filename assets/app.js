$(document).ready(function() {
  var linesToFetch = 10;
  var needle = $('section.results').data('needle');
  $('div.channel').highlight(needle);

  var colors = ["rgb(204,102,102)", "rgb(173,104,0)", "rgb(56,112,95)", "rgb(43,95,173)", "rgb(121,36,143)", "rgb(143,36,36)", "rgb(204,189,51)", "rgb(0,173,121)", "rgb(0,45,112)", "rgb(163,0,204)", "rgb(204,0,0)", "rgb(112,101,0)", "rgb(51,204,204)", "rgb(0,82,204)", "rgb(204,102,163)", "rgb(143,93,71)", "rgb(129,143,71)", "rgb(56,95,112)", "rgb(22,0,112)", "rgb(112,28,79)", "rgb(204,97,51)", "rgb(139,173,0)", "rgb(102,173,204)", "rgb(84,56,112)", "rgb(173,0,104)", "rgb(143,43,0)", "rgb(69,173,43)", "rgb(0,121,173)", "rgb(87,0,173)", "rgb(143,71,93)", "rgb(143,114,71)", "rgb(28,112,36)", "rgb(56,79,112)", "rgb(184,102,204)", "rgb(204,0,61)"];
  var cur_color = 0;
  var sender_colors = {};

  function colorize_senders(within_elem) {
    var sender;
    $(within_elem).find('.sender').each(function (idx, elem) {
      sender = $(elem).text();
      if (!(sender in sender_colors)) {
        sender_colors[sender] = colors[cur_color++];
      }
      $(elem).css('color', sender_colors[sender]);
    });
  }
  colorize_senders($('section'));

  // compares 2 objects (e.g. in a sort call) using the specified key
  function cmp_key(key) {
    return function(a,b) {
      a = a[key], b = b[key]
      if (a < b)
        return -1;
      else if (a > b)
        return 1;
      else
        return 0;
    }
  }

  function getAndInsertContext(messageId, linesToFetch, direction) {
    console.log(this);
    $.ajax({
      url: "/context/",
      data: {messageId: messageId,
             linesToFetch: linesToFetch,
             direction: direction},
      context: this,
    }).done(function(messages) {
      // sort messages ASC if aftercontext, DESC if beforecontext
      messages = messages.sort(cmp_key("MessageId"));
      if (direction == -1)
        messages.reverse()

      var msgElem, msg, dt;
      var directionSelect = (direction == -1) ? ".before" : ".after";
      var directionFunc = (direction == -1) ? "prependTo" : "appendTo";
      for (var i = 0; i < messages.length; i++) {
        msg = messages[i];
        dt = new Date(msg.Time);
        msgElem = $(this).siblings('.matching-line').children('.message').first().clone()
        //$(msgElem).data('messageid', msg.MessageId);
        $(msgElem).attr('data-messageid', msg.MessageId);
        $(msgElem).children('.time').text(dt.toLocaleString());
        $(msgElem).children('.sender').text(msg.Sender.Username);
        $(msgElem).children('.sender').attr('title', msg.Sender.FullIdent);
        $(msgElem).children('.text').text(msg.Text);
        $(msgElem)[directionFunc](this);
      }
      $(this).highlight(needle);
      colorize_senders($(this));
    });
  }

  //$('.matching-line').click(function() {
    //var lowMsgId = $(this).siblings('.before.context').children().first().data('messageid') ||
                   //$(this).children('.message').data('messageid');
    //var highMsgId = $(this).siblings('.after.context').children().last().data('messageid') ||
                    //$(this).children('.message').data('messageid');
    //getAndInsertContext.call(this, lowMsgId, linesToFetch, -1);
    //getAndInsertContext.call(this, highMsgId, linesToFetch, 1);
  //});
  $('.before.expand').click(function() {
    var beforeMsgs = $(this).siblings('.before.context');
    var messageId = $(beforeMsgs).children().first().data('messageid') ||
                    $(this).siblings('.matching-line').children('.message').data('messageid');
    getAndInsertContext.call(beforeMsgs, messageId, linesToFetch, -1);
  });
  $('.after.expand').click(function() {
    var afterMsgs = $(this).siblings('.after.context');
    var messageId = $(afterMsgs).children().last().data('messageid') ||
                    $(this).siblings('.matching-line').children('.message').data('messageid');
    getAndInsertContext.call(afterMsgs, messageId, linesToFetch, 1);
  });
  $('.results h4').click(function() {
    $(this).next('.channel').slideToggle(200);
    $(this).toggleClass("collapsed");
  });
});
