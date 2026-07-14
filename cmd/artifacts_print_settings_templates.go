package cmd

// Snippet templates used by `artifacts print-settings` subcommands. They are
// direct ports of the templates from gcloud-python's
// googlecloudsdk/command_lib/artifacts/print_settings/*.py so that the emitted
// snippets are byte-for-byte compatible.

const gradleServiceAccountTemplate = `// Move the secret to ~/.gradle.properties
def artifactRegistryMavenSecret = "{password}"

// Insert following snippet into your build.gradle
// see docs.gradle.org/current/userguide/publishing_maven.html

plugins {
  id "maven-publish"
}

publishing {
  repositories {
    maven {
      url "https://{location}-maven.pkg.dev/{repo_path}"
      credentials {
        username = "{username}"
        password = "$artifactRegistryMavenSecret"
      }
    }
  }
}

repositories {
  maven {
    url "https://{location}-maven.pkg.dev/{repo_path}"
    credentials {
      username = "{username}"
      password = "$artifactRegistryMavenSecret"
    }
    authentication {
      basic(BasicAuthentication)
    }
  }
}
`

const gradleServiceAccountSnapshotTemplate = `// Move the secret to ~/.gradle.properties
def artifactRegistryMavenSecret = "{password}"

// Insert following snippet into your build.gradle
// see docs.gradle.org/current/userguide/publishing_maven.html

plugins {
  id "maven-publish"
}

publishing {
  repositories {
    maven {
      def snapshotURL = "https://{location}-maven.pkg.dev/{repo_path}"
      def releaseURL = "<Paste release URL here>"
      url version.endsWith('SNAPSHOT') ? snapshotURL : releaseURL
      credentials {
        username = "{username}"
        password = "$artifactRegistryMavenSecret"
      }
    }
  }
}

repositories {
  maven {
    url "https://{location}-maven.pkg.dev/{repo_path}"
    credentials {
      username = "{username}"
      password = "$artifactRegistryMavenSecret"
    }
    authentication {
      basic(BasicAuthentication)
    }
  }
}
`

const gradleServiceAccountReleaseTemplate = `// Move the secret to ~/.gradle.properties
def artifactRegistryMavenSecret = "{password}"

// Insert following snippet into your build.gradle
// see docs.gradle.org/current/userguide/publishing_maven.html

plugins {
  id "maven-publish"
}

publishing {
  repositories {
    maven {
      def snapshotURL = "<Paste snapshot URL here>"
      def releaseURL = "https://{location}-maven.pkg.dev/{repo_path}"
      url version.endsWith('SNAPSHOT') ? snapshotURL : releaseURL
      credentials {
        username = "{username}"
        password = "$artifactRegistryMavenSecret"
      }
    }
  }
}

repositories {
  maven {
    url "https://{location}-maven.pkg.dev/{repo_path}"
    credentials {
      username = "{username}"
      password = "$artifactRegistryMavenSecret"
    }
    authentication {
      basic(BasicAuthentication)
    }
  }
}
`

const gradleNoServiceAccountTemplate = `// Insert following snippet into your build.gradle
// see docs.gradle.org/current/userguide/publishing_maven.html

plugins {
  id "maven-publish"
  id "com.google.cloud.artifactregistry.gradle-plugin" version "{extension_version}"
}

publishing {
  repositories {
    maven {
      url "artifactregistry://{location}-maven.pkg.dev/{repo_path}"
    }
  }
}

repositories {
  maven {
    url "artifactregistry://{location}-maven.pkg.dev/{repo_path}"
  }
}
`

const gradleNoServiceAccountSnapshotTemplate = `// Insert following snippet into your build.gradle
// see docs.gradle.org/current/userguide/publishing_maven.html

plugins {
  id "maven-publish"
  id "com.google.cloud.artifactregistry.gradle-plugin" version "{extension_version}"
}

publishing {
  repositories {
    maven {
      def snapshotURL = "artifactregistry://{location}-maven.pkg.dev/{repo_path}"
      def releaseURL = "<Paste release URL here>"
      url version.endsWith('SNAPSHOT') ? snapshotURL : releaseURL
    }
  }
}

repositories {
  maven {
    url "artifactregistry://{location}-maven.pkg.dev/{repo_path}"
  }
}
`

const gradleNoServiceAccountReleaseTemplate = `// Insert following snippet into your build.gradle
// see docs.gradle.org/current/userguide/publishing_maven.html

plugins {
  id "maven-publish"
  id "com.google.cloud.artifactregistry.gradle-plugin" version "{extension_version}"
}

publishing {
  repositories {
    maven {
      def snapshotURL = "<Paste snapshot URL here>"
      def releaseURL = "artifactregistry://{location}-maven.pkg.dev/{repo_path}"
      url version.endsWith('SNAPSHOT') ? snapshotURL : releaseURL
    }
  }
}

repositories {
  maven {
    url "artifactregistry://{location}-maven.pkg.dev/{repo_path}"
  }
}
`

const mvnServiceAccountTemplate = `<!-- Insert following snippet into your pom.xml -->

<project>
  <distributionManagement>
    <snapshotRepository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
    </snapshotRepository>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
    </repository>
  </distributionManagement>

  <repositories>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
      <releases>
        <enabled>true</enabled>
      </releases>
      <snapshots>
        <enabled>true</enabled>
      </snapshots>
    </repository>
  </repositories>
</project>

<!-- Insert following snippet into your settings.xml -->

<settings>
  <servers>
    <server>
      <id>{server_id}</id>
      <configuration>
        <httpConfiguration>
          <get>
            <usePreemptive>true</usePreemptive>
          </get>
          <head>
            <usePreemptive>true</usePreemptive>
          </head>
          <put>
            <params>
              <property>
                <name>http.protocol.expect-continue</name>
                <value>false</value>
              </property>
            </params>
          </put>
        </httpConfiguration>
      </configuration>
      <username>{username}</username>
      <password>{password}</password>
    </server>
  </servers>
</settings>
`

const mvnNoServiceAccountTemplate = `<!-- Insert following snippet into your pom.xml -->

<project>
  <distributionManagement>
    <snapshotRepository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
    </snapshotRepository>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
    </repository>
  </distributionManagement>

  <repositories>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
      <releases>
        <enabled>true</enabled>
      </releases>
      <snapshots>
        <enabled>true</enabled>
      </snapshots>
    </repository>
  </repositories>

  <build>
    <extensions>
      <extension>
        <groupId>com.google.cloud.artifactregistry</groupId>
        <artifactId>artifactregistry-maven-wagon</artifactId>
        <version>{extension_version}</version>
      </extension>
    </extensions>
  </build>
</project>
`

const mvnNoServiceAccountSnapshotTemplate = `<!-- Insert following snippet into your pom.xml -->

<project>
  <distributionManagement>
    <snapshotRepository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
    </snapshotRepository>
  </distributionManagement>

  <repositories>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
      <releases>
        <enabled>false</enabled>
      </releases>
      <snapshots>
        <enabled>true</enabled>
      </snapshots>
    </repository>
  </repositories>

  <build>
    <extensions>
      <extension>
        <groupId>com.google.cloud.artifactregistry</groupId>
        <artifactId>artifactregistry-maven-wagon</artifactId>
        <version>{extension_version}</version>
      </extension>
    </extensions>
  </build>
</project>
`

const mvnNoServiceAccountReleaseTemplate = `<!-- Insert following snippet into your pom.xml -->

<project>
  <distributionManagement>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
    </repository>
  </distributionManagement>

  <repositories>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
      <releases>
        <enabled>true</enabled>
      </releases>
      <snapshots>
        <enabled>false</enabled>
      </snapshots>
    </repository>
  </repositories>

  <build>
    <extensions>
      <extension>
        <groupId>com.google.cloud.artifactregistry</groupId>
        <artifactId>artifactregistry-maven-wagon</artifactId>
        <version>{extension_version}</version>
      </extension>
    </extensions>
  </build>
</project>
`

const mvnServiceAccountSnapshotTemplate = `<!-- Insert following snippet into your pom.xml -->

<project>
  <distributionManagement>
    <snapshotRepository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
    </snapshotRepository>
  </distributionManagement>

  <repositories>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
      <releases>
        <enabled>false</enabled>
      </releases>
      <snapshots>
        <enabled>true</enabled>
      </snapshots>
    </repository>
  </repositories>
</project>

<!-- Insert following snippet into your settings.xml -->

<settings>
  <servers>
    <server>
      <id>{server_id}</id>
      <configuration>
        <httpConfiguration>
          <get>
            <usePreemptive>true</usePreemptive>
          </get>
          <head>
            <usePreemptive>true</usePreemptive>
          </head>
          <put>
            <params>
              <property>
                <name>http.protocol.expect-continue</name>
                <value>false</value>
              </property>
            </params>
          </put>
        </httpConfiguration>
      </configuration>
      <username>{username}</username>
      <password>{password}</password>
    </server>
  </servers>
</settings>
`

const mvnServiceAccountReleaseTemplate = `<!-- Insert following snippet into your pom.xml -->

<project>
  <distributionManagement>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
    </repository>
  </distributionManagement>

  <repositories>
    <repository>
      <id>{server_id}</id>
      <url>{scheme}://{location}-maven.pkg.dev/{repo_path}</url>
      <releases>
        <enabled>true</enabled>
      </releases>
      <snapshots>
        <enabled>false</enabled>
      </snapshots>
    </repository>
  </repositories>
</project>

<!-- Insert following snippet into your settings.xml -->

<settings>
  <servers>
    <server>
      <id>{server_id}</id>
      <configuration>
        <httpConfiguration>
          <get>
            <usePreemptive>true</usePreemptive>
          </get>
          <head>
            <usePreemptive>true</usePreemptive>
          </head>
          <put>
            <params>
              <property>
                <name>http.protocol.expect-continue</name>
                <value>false</value>
              </property>
            </params>
          </put>
        </httpConfiguration>
      </configuration>
      <username>{username}</username>
      <password>{password}</password>
    </server>
  </servers>
</settings>
`

const npmServiceAccountTemplate = `# Insert the following snippet into your project .npmrc

{configured_registry}=https://{registry_path}
//{registry_path}:always-auth=true

# Insert the following snippet into your user .npmrc

//{registry_path}:_password="{password}"
//{registry_path}:username=_json_key_base64
//{registry_path}:email=not.valid@email.com
`

const npmNoServiceAccountTemplate = `# Insert the following snippet into your project .npmrc

{configured_registry}=https://{registry_path}
//{registry_path}:always-auth=true
`

const pythonServiceAccountTemplate = `# Insert the following snippet into your .pypirc

[distutils]
index-servers =
    {repo}

[{repo}]
repository: https://{location}-python.pkg.dev/{repo_path}/
username: _json_key_base64
password: {password}

# Insert the following snippet into your pip.conf

[global]
extra-index-url = https://_json_key_base64:{password}@{location}-python.pkg.dev/{repo_path}/simple/
`

const pythonNoServiceAccountTemplate = `# Insert the following snippet into your .pypirc

[distutils]
index-servers =
    {repo}

[{repo}]
repository: https://{location}-python.pkg.dev/{repo_path}/

# Insert the following snippet into your pip.conf

[global]
extra-index-url = https://{location}-python.pkg.dev/{repo_path}/simple/
`
