package xen_xl

var defaultMetdataPVHVM = []byte(`builder = 'hvm'
name = '{{.DomainName}}'
memory = {{.RAM}}
vcpus = {{.CPUs}}
acpi = 1 
apic = 1
{{.Networks}} 
{{.Storage}}
boot = 'd'
xen_platform_pci = 1
on_poweroff = 'destroy'
on_reboot   = 'restart'
on_crash    = 'restart'
sdl = 1
vnc = 0
vncpasswd = ''
stdvga = 0
serial = 'pty'
tsc_mode = 'default'
`)
