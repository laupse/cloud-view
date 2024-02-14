import asyncio
import sys
import dagger
import click
from functools import wraps

GO_VERSION = "1.22.0"
ALPINE_VERSION = "3.19.0"


async def publish_image(
    dagger_client: dagger.Client,
    tag: str,
    registry_address: str,
    registry_username: str,
    registry_password: str,
):
    final = await build_image(dagger_client)
    if registry_username is not None:
        password = dagger_client.set_secret("registry_password", registry_password)
        final = final.with_registry_auth(registry_address, registry_username, password)
    await final.publish(f"{registry_address}:{tag}")


async def build_image(dagger_client: dagger.Client) -> dagger.Container:
    gomod = dagger_client.host().file("./go.mod")
    gosum = dagger_client.host().file("./go.sum")
    views = dagger_client.host().directory("./internal/views", exclude=["_templ.go"])
    src = dagger_client.host().directory(
        ".", include=["**/*.go"], exclude=["_templ.go"]
    )
    static = dagger_client.host().directory("./static")

    tailwind_config = dagger_client.host().file("./tailwind.config.js")
    # css = dagger_client.host().directory("./internal/views/css")

    dist = (
        dagger_client.container()
        .from_(f"alpine:{ALPINE_VERSION}")
        .with_exec(["apk", "add", "curl"])
        .with_exec(
            [
                "curl",
                "-sLO",
                "https://github.com/tailwindlabs/tailwindcss/releases/download/v3.3.6/tailwindcss-linux-x64",
            ]
        )
        .with_exec(["chmod", "+x", "tailwindcss-linux-x64"])
        .with_file("tailwind.config.js", tailwind_config)
        .with_mounted_directory("./internal/views/", views)
        .with_exec(
            [
                "./tailwindcss-linux-x64",
                "-i",
                "./internal/views/css/input.css",
                "-o",
                "dist/output.css",
            ]
        )
        .directory("dist")
    )
    golang_image = (
        dagger_client.container().from_(f"golang:{GO_VERSION}").with_workdir("/src")
    )

    dep = (
        golang_image.with_mounted_file("go.mod", gomod)
        .with_mounted_file("go.sum", gosum)
        .with_exec(["go", "mod", "download", "-x"])
    )

    template = (
        golang_image.with_mounted_directory("internal/views", views)
        .with_exec(["go", "install", "github.com/a-h/templ/cmd/templ@latest"])
        .with_exec(["templ", "generate"])
        .directory("internal/views")
    )

    bin = (
        dep.with_directory(".", src)
        .with_directory("./internal/views", template)
        .with_env_variable("CGO_ENABLED", "0")
        .with_exec(["go", "build", "-o", "/bin/cloud-view"])
        .file("/bin/cloud-view")
    )

    final = (
        dagger_client.container()
        .from_(f"alpine:{ALPINE_VERSION}")
        .with_workdir("/bin")
        .with_directory("static", static)
        .with_directory("dist", dist)
        .with_file("cloud-view", bin)
        .with_entrypoint(["."])
    )

    return final


def coro(f):
    @wraps(f)
    def wrapper(*args, **kwargs):
        return asyncio.run(f(*args, **kwargs))

    return wrapper


@click.group()
def cli():
    pass


@cli.command()
@coro
@click.option("--tag", required=True)
@click.option("--registry-address", required=True, envvar="REGISTRY_ADDRESS")
@click.option("--registry-username", default="", envvar="REGISTRY_USERNAME")
@click.option("--registry-password", default="", envvar="REGISTRY_PASSWORD")
async def publish(tag, registry_address, registry_username, registry_password):
    config = dagger.Config(log_output=sys.stdout)
    async with dagger.Connection(config) as client:
        await publish_image(
            client, tag, registry_address, registry_username, registry_password
        )


@cli.command()
@coro
async def build():
    config = dagger.Config(log_output=sys.stdout)
    async with dagger.Connection(config) as client:
        build = await build_image(client)
        await build.export("cloud-view.tar.gz")


if __name__ == "__main__":
    cli()
