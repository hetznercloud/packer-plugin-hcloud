## 1.0.2 (November 12, 2021)

* Add pre validate step check for snapshot names. Invoking packer build -force
    will bypass the checks, thus multiple snapshots with the same name will be
    allowed. This is equivalent to the previous behavior of this plugin.
    [GH-17] [GH-18]
* Upgrade packer-plugin-sdk to v0.2.9. [GH-23]

## 1.0.1 (September 1, 2021)

* Add SSH key support for the Freebsd64 Rescue System. [GH-9]
* Bump to Go 1.17

## 1.0.0 (June 14, 2021)
* upgrade packer-plugin-sdk to v0.2.3. [GH-8]

## 0.0.1 (April 22, 2021)
* initial extraction of hcloud plugin from Packer core.
