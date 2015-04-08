function! Run(cmd)
  let cmdStr = "silent !runner -cmd " . a:cmd . " > /dev/null"
  execute cmdStr
  redraw!
endfunction
