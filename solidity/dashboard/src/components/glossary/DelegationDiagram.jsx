import React from "react"
import Tile from "../Tile"
import * as Icons from "../Icons"

const roles = [
  {
    icon: <Icons.Authorizer />,
    name: "authorizer",
    description:
      "A role that approves operator contracts and slashing rules for operator misbehavior.",
  },
  {
    icon: <Icons.Operations />,
    name: "operator",
    description:
      "The operator address is tasked with participation in network operations, and represents the staker in most circumstances.",
  },
  {
    icon: <Icons.Rewards />,
    name: "beneficiary",
    description:
      "The address to which rewards are sent that are generated by stake doing work on the network.",
  },
]

const DelegationDiagram = () => {
  return (
    <Tile title="Diagram of Delegation Roles" titleClassName="h3 text-grey-70">
      <div className="flex row center mt-2">
        <Icons.KeepOutline />
        <h4 className="ml-1">Owner</h4>
      </div>
      <div className="text-big text-grey-60" style={{ marginTop: "0.5rem" }}>
        The original KEEP token holder. The owner delegates stake to a trusted
        third party (operator) to stake on their behalf.
      </div>
      <section id="delegation-diagram">
        <section className="roles">
          <ul>{roles.map(renderRole)}</ul>
        </section>
        <section className="diagram">
          <Icons.DelegationDiagram />
        </section>
      </section>
    </Tile>
  )
}

export default DelegationDiagram

const renderRole = (role) => <Role key={role.name} {...role} />

const Role = ({ icon, name, description }) => {
  return (
    <li>
      <div className="flex row center">
        {icon}
        <h5 className="text-grey-60 ml-1">{name}</h5>
      </div>
      <div className="text-small text-grey-40">{description}</div>
    </li>
  )
}
