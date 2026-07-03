<!-- page:004 -->

```common-lisp
(defmethod present ((self box) stream)
  (draw-border self stream))
```

Empty code blocks are skipped entirely.
