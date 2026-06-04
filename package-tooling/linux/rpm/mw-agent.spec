Name: %{package_name}
Version: %{release_version}
Release: %{release_number}
Summary: Middleware Agent(%{package_name}) Service
License: GPL
Group: Development/Tools
Source0: %{package_name}-%{release_version}-%{arch}.tar.gz
Provides: %{package_name}
Obsoletes: %{package_name} < %{release_version}

%description
Middleware Agent(%{package_name}) service enables you to monitor your infrastructure and applications.

%global __strip /bin/true

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/opt/%{package_name}/bin
mkdir -p %{buildroot}/etc/%{package_name}
mkdir -p %{buildroot}/lib/systemd/system
mkdir -p %{buildroot}/etc/%{package_name}/opamp

cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/%{package_name} %{buildroot}/opt/%{package_name}/bin/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/mw-opamp-supervisor %{buildroot}/opt/%{package_name}/bin/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/agent-config.yaml.sample %{buildroot}/etc/%{package_name}/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/mw-agent.env.sample %{buildroot}/etc/%{package_name}/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/otel-config.yaml.sample %{buildroot}/etc/%{package_name}/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/supervisor-config.yaml.sample %{buildroot}/etc/%{package_name}/
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/postinstall.sh %{buildroot}/opt/%{package_name}/.postinstall.sh
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/%{package_name}.service %{buildroot}/lib/systemd/system/%{package_name}.service
cp -rfa ~/build/rpmbuild/SOURCES/%{arch}/%{package_name}-%{release_version}/%{package_name}-opamp.service %{buildroot}/lib/systemd/system/%{package_name}-opamp.service


%files
/opt/%{package_name}/bin/%{package_name}
/opt/%{package_name}/bin/mw-opamp-supervisor
/etc/%{package_name}/agent-config.yaml.sample
/etc/%{package_name}/mw-agent.env.sample
/etc/%{package_name}/otel-config.yaml.sample
/etc/%{package_name}/supervisor-config.yaml.sample
/opt/%{package_name}/.postinstall.sh
/lib/systemd/system/%{package_name}.service
/lib/systemd/system/%{package_name}-opamp.service

%post
chmod u+x /opt/%{package_name}/.postinstall.sh
/opt/%{package_name}/.postinstall.sh

%preun
if [ $1 -gt 0 ]; then
    echo "Upgrade in progress, skipping pre-uninstallation steps."
else
    # Stop and disable mw-agent service if running
    if systemctl is-active --quiet %{package_name} 2>/dev/null; then
        systemctl stop %{package_name}
    fi
    
    if systemctl is-enabled --quiet %{package_name} 2>/dev/null; then
        systemctl disable %{package_name}
    fi
    
    # Stop and disable mw-agent-opamp service if running
    if systemctl is-active --quiet %{package_name}-opamp 2>/dev/null; then
        systemctl stop %{package_name}-opamp
    fi
    
    if systemctl is-enabled --quiet %{package_name}-opamp 2>/dev/null; then
        systemctl disable %{package_name}-opamp
    fi
fi

%postun
if [ $1 -gt 0 ]; then
    echo "Upgrade in progress, skipping post-uninstallation steps."
else
    rm -rf /etc/%{package_name}
    # rmdir  /etc/%{package_name}
    rmdir /opt/%{package_name}/bin
    rmdir /opt/%{package_name}
fi

