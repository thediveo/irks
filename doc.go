/*
Package irks provides information about IRQ counts per CPU on Linux, as well as
some IRQ “topology information”.

# The Format of /proc/interrupts

Unfortunately, the man page for [proc_interrupts(5)] has not much to say, except
“Very easy to read formatting, done in ASCII.” This “explanation” almost comical
in a true Hitchhiker's Guide way of fashion.

Digging into the Linux kernel sources luckily brings up [show_interrupts] that
then spills the beans about the format of the “/proc/interrupts” pseudo file.

First comes the header line...

  - space padding covering the width of the following IRQ number columns.
  - then, for each CPU that currently is online, the string "CPU" followed by
    the CPU number and column padding.

Then come the individual IRQ lines...

  - right aligned, space-padded IRQ number, followed by “:” and a trailing single
    space. Please note that some of the trailing interrupt lines do not have IRQ
    numbers, but names instead, as they are “[architecture-specific interrupts]”.
  - then, for each CPU that currently is online, the count, right aligned,
    space-padded, of width 10, and a single trailing space.
  - information about the IRQ chip involved, if available (otherwise a padded
    “None”); in the worst case, this is free-style text registered by some kernel
    board driver stuff.
  - if available, IRQ domain information.
  - if the kernel was compiled with the particular option, then the generic type
    of IRQ trigger, either “Level” or “Edge”.
  - the IRQ descriptive name, if set, otherwise this information is simply left
    out. If it contains spaces ... you get the spaces ... 🤷
  - if “actions” are assigned to this IRQ, then two spaces follow, and then the
    list of actions, separated each by “, ”. However, this information is much
    easier glanced from “/sys/kernel/irq/#/actions” (see next).

# The Format of /sys/kernel/irq/#/

Information about individual IRQs is also available in a second place, but
compared to “/proc/interrupts” now broken up into many individual data tidbits
instead of a single pseudo file. The first level is per IRQ number, hence the
metasyntactic “#” in “/sys/kernel/irq/#/”. For each IRQ there is a set of
individual pseudo files, please see also the [kernel ABI testing documentation
on /sys/kernel/irq]:

  - “actions”: the IRQ action chain in form of a comma-separated list of zero or
    more actions associated with this interrupt. Actions might be device names,
    but also other elements, such as individual RX/TX queue IRQs of network cards.
  - “chip_name”
  - “hwirq”
  - “name”: the clear-text name of the flow handler, such as “edge”, et cetera.
  - “per_cpu_count”: a list of comma-separated counters per CPU that currently is
    in the system, either online of offline. This field thus differs from
    “/proc/interrupts”, where the latter only lists CPUs that are currently online.
  - “type”: either “edge” or “level”.
  - “wakeup”: wakeup state of interrupt, either “enabled” or “disabled”.

The downside of “/sys/kernel/irq/#/” is that gathering all information requires
a lot of repeated open, read, and close VFS operations. In contrast, getting the
IRQ counters per CPU requires considerably fewer VFS operations when using
“/proc/interrupts”: one open, one close, and inbetween just reading, reading,
reading. From a performance perspective, “/sys/kernel/irq/#/” should be used in
order to get certain structural IRQ information, such as the actions.

# The Format of /proc/irq/#/

Oh, there's a third place that also provides further IRQ information. Its main
function is to show and control the IRQ-to-CPU(s) affinities.

  - “affinity_hint”
  - “effective_affinity”
  - “effective_affinity_list”
  - “node”
  - “smp_affinity”
  - “smp_affinity_list”
  - “spurious”
  - “$HANDLER” ([register_handler_proc])

[proc_interrupts(5)]: https://man7.org/linux/man-pages/man5/proc_interrupts.5.html
[show_interrupts]: https://elixir.bootlin.com/linux/v6.12/source/kernel/irq/proc.c#L463
[architecture-specific interrupts]: https://elixir.bootlin.com/linux/v6.12/source/arch/x86/kernel/irq.c#L61
[register_handler_proc]: https://elixir.bootlin.com/linux/v6.12/source/kernel/irq/proc.c#L317
[kernel ABI testing documentation on /sys/kernel/irq]: https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-kernel-irq
*/
package irks
