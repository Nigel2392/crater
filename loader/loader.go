package loader

import (
	"fmt"
	"syscall/js"
)

type (
	Loader        []interface{}
	ClassList     []string
	Style         map[string]string
	ID            string
	Element       js.Value
	OnShowFunc    func(Loader)
	OnHideFunc    func(Loader)
	QuerySelector string
)

func NewLoader(args ...any) (Loader, error) {
	var loader = make(Loader, 0, len(args))
	var argMap = make(map[string]struct{})
	for _, arg := range args {
		switch arg := arg.(type) {
		case ClassList:
			if _, ok := argMap["classList"]; ok {
				return nil, fmt.Errorf("duplicate class %s", arg)
			}
			loader = append(loader, arg)
			argMap["classList"] = struct{}{}
		case Style:
			if _, ok := argMap["style"]; ok {
				return nil, fmt.Errorf("duplicate style %s", arg)
			}
			loader = append(loader, arg)
			argMap["style"] = struct{}{}
		case ID:
			if _, ok := argMap["id"]; ok {
				return nil, fmt.Errorf("duplicate id %s", arg)
			}
			loader = append(loader, arg)
			argMap["id"] = struct{}{}
		case Element:
			if _, ok := argMap["element"]; ok {
				return nil, fmt.Errorf("duplicate element %v", arg)
			}
			loader = append(loader, arg)
			argMap["element"] = struct{}{}
		case OnShowFunc:
			if _, ok := argMap["onShow"]; ok {
				return nil, fmt.Errorf("duplicate onShow %v", arg)
			}
			loader = append(loader, arg)
			argMap["onShow"] = struct{}{}
		case OnHideFunc:
			if _, ok := argMap["onHide"]; ok {
				return nil, fmt.Errorf("duplicate onHide %v", arg)
			}
			loader = append(loader, arg)
			argMap["onHide"] = struct{}{}
		case QuerySelector:
			if _, ok := argMap["querySelector"]; ok {
				return nil, fmt.Errorf("duplicate querySelector %v", arg)
			}
			loader = append(loader, arg)
			argMap["querySelector"] = struct{}{}
		default:
			return nil, fmt.Errorf("invalid type %T", arg)
		}
	}
	return loader, nil
}

type loaderOpts struct {
	element       js.Value
	style         map[string]string
	class         []string
	id            string
	onShow        OnShowFunc
	onHide        OnHideFunc
	querySelector string
}

func (l Loader) opts() loaderOpts {
	var element js.Value
	var style map[string]string
	var classList []string
	var id string
	var onShow func(Loader)
	var onHide func(Loader)
	var querySelector string
	for _, arg := range l {
		switch arg := arg.(type) {
		case Element:
			element = js.Value(arg)
		case Style:
			style = arg
		case ClassList:
			classList = arg
		case ID:
			id = string(arg)
		case OnShowFunc:
			onShow = arg
		case OnHideFunc:
			onHide = arg
		case QuerySelector:
			querySelector = string(arg)
		default:
			panic(fmt.Errorf("invalid type %T", arg))
		}
	}
	var e js.Value
	if !element.IsUndefined() && !element.IsNull() {
		e = element
	} else if querySelector != "" {
		element = js.Global().Get("document").Call("querySelector", querySelector)
		e = element
	} else {
		e = js.Global().Get("document").Get("body")
	}
	return loaderOpts{
		element: e,
		style:   style,
		class:   classList,
		id:      id,
		onShow:  onShow,
		onHide:  onHide,
	}
}

func (l Loader) Show() {
	var opts = l.opts()

	if opts.onShow != nil {
		opts.onShow(l)
	}

	var element = opts.element
	if element.IsUndefined() || element.IsNull() {
		return
	}
	var style = element.Get("style")
	var classList = element.Get("classList")
	for k, v := range opts.style {
		style.Set(k, v)
	}
	if len(opts.class) > 0 {
		for _, class := range opts.class {
			classList.Call("add", class)
		}
	}
	if opts.id != "" {
		opts.element.Set("id", opts.id)
	}
}

func (l Loader) Hide() {
	var opts = l.opts()

	if opts.onHide != nil {
		opts.onHide(l)
	}

	var element = opts.element
	if element.IsUndefined() || element.IsNull() {
		return
	}
	var style = element.Get("style")
	var classList = element.Get("classList")
	for k := range opts.style {
		style.Delete(k)
	}
	if len(opts.class) > 0 {
		for _, class := range opts.class {
			classList.Call("remove", class)
		}
	}
	if opts.id != "" {
		opts.element.Delete("id")
	}
}
